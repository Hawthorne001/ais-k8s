#!/usr/bin/env bash
set -e

#### DO NOT USE FOR PRODUCTION ####

# This script
# 1. Creates a kind cluster
# 2. Installs all necessary prerequisites for keycloak
# 3. Installs keycloak
# 4. Imports AIS realm

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLUSTER_NAME="keycloak-test"
KEYCLOAK_VERSION="26.6.0"
LOCAL_SCRIPT_DIR="${SCRIPT_DIR}/../../local"

if ! command -v yq >/dev/null 2>&1; then
  echo "ERROR: yq is required to import the AIS realm (used by realm/import-realm.sh)" >&2
  exit 1
fi

source "${LOCAL_SCRIPT_DIR}/start-kind.sh"
create_kind_cluster $CLUSTER_NAME

# Install pre-reqs -- storage class, ingress etc.
helmfile -f prereq-helmfile.yaml sync
# Create namespace but allow existing
kubectl create namespace keycloak || true

# Install cluster issuer for making cert
helmfile -f ../../helm/issuer/helmfile.yaml sync
# Create a certificate
kubectl apply -f manifests/certificate.yaml

# Install cnpg
helmfile -f ./cnpg/helm/operator/helmfile.yaml sync
echo "Waiting for cnpg controller webhook to become available..."
kubectl rollout status deployment/cloudnative-pg-operator -n cnpg-system --timeout=120s
helmfile -f ./cnpg/helm/cluster/helmfile.yaml sync

#### KEYCLOAK ####
# CRDs
kubectl apply -f "https://raw.githubusercontent.com/keycloak/keycloak-k8s-resources/${KEYCLOAK_VERSION}/kubernetes/keycloaks.k8s.keycloak.org-v1.yml"
kubectl apply -f "https://raw.githubusercontent.com/keycloak/keycloak-k8s-resources/${KEYCLOAK_VERSION}/kubernetes/keycloakrealmimports.k8s.keycloak.org-v1.yml"

# Operator
kubectl -n keycloak apply -f "https://raw.githubusercontent.com/keycloak/keycloak-k8s-resources/${KEYCLOAK_VERSION}/kubernetes/kubernetes.yml"

# Create a secret for keycloak to access the DB, allow existing
kubectl create secret -n keycloak generic keycloak-db-secret --from-literal=username=app --from-literal=password="$(kubectl get secret cloudnative-pg-cluster-app --namespace cnpg-database -o jsonpath='{.data.password}' | base64 --decode)" || true

# Manifest
kubectl apply -f manifests/keycloak.yaml
until kubectl get keycloak keycloak-server -n keycloak; do
  echo "Waiting for keycloak-server custom resource to exist..."
  sleep 5
done
echo "Waiting for keycloak to be ready (takes some time)..."
kubectl wait --for=condition=Ready --timeout=180s keycloak/keycloak-server -n keycloak

# TODO: Run this only when necessary
# Run import realm job
./realm/import-realm.sh

# Print initial temp admin credentials
USER=$(kubectl get secret -n keycloak keycloak-server-initial-admin -o jsonpath='{.data.username}' | base64 --decode)
PASS=$(kubectl get secret -n keycloak keycloak-server-initial-admin -o jsonpath='{.data.password}' | base64 --decode)

# Start a port forward and kill at the end of the script
kubectl port-forward -n keycloak service/keycloak-server-service 8543:8543 >/dev/null 2>&1 &
pid=$!
trap "kill $pid" EXIT

# kubectl port-forward binds asynchronously, so wait for the listener before using it
deadline=$((SECONDS + 30))
until (exec 3<>/dev/tcp/localhost/8543) 2>/dev/null; do
  if ! kill -0 "$pid" 2>/dev/null; then
    echo "ERROR: port-forward to keycloak-server-service exited" >&2
    exit 1
  fi
  if ((SECONDS >= deadline)); then
    echo "ERROR: timed out waiting for port-forward on localhost:8543" >&2
    exit 1
  fi
  echo "Waiting for keycloak port-forward..."
  sleep 1
done

# Get ca.crt for trust from the issuer
CA_FILE=$SCRIPT_DIR/scripts/ca.crt
kubectl get secret ca-root-secret -n cert-manager -o "jsonpath={.data['ca\.crt']}" | base64 -d > "$CA_FILE"
# Create an ais-admin user through the port-forward above.
# The certificate includes a localhost SAN, so this needs no hosts entry for the internal service name.
KEYCLOAK_HOST="https://localhost:8543"
echo "$PASS" | "$SCRIPT_DIR/scripts/prepare_cluster.sh" "$KEYCLOAK_HOST" "$USER" "$CA_FILE"

echo ""
echo "Initial admin user: ${USER}"
echo "Initial admin password: ${PASS}"
echo ""
echo "To access locally, port forward the keycloak service 'kubectl port-forward -n keycloak service/keycloak-server-service 8543:8543'"
echo "View the AIStore realm:"
echo "curl -k https://localhost:8543/realms/aistore"
