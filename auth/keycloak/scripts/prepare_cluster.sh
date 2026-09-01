#!/usr/bin/env bash
set -e

# This script creates an ais-admin user in an active keycloak
# Note this REQUIRES a port-forward to already be running or a locally accessible cluster
# The keycloak admin password is prompted, or read from stdin when piped in

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
  echo "Usage: $0 <HOST> <USER> [CA_CRT_PATH]" >&2
  exit 1
}

# Require at least 2 args, 3rd is optional
if [ "$#" -lt 2 ] || [ "$#" -gt 3 ]; then
  usage
fi

KEYCLOAK_HOST="$1"
USER="$2"
CA_FILE="${3:-}"

if [ -n "$CA_FILE" ] && [ ! -f "$CA_FILE" ]; then
  echo "CA cert '$CA_FILE' not found" >&2
  usage
fi

# Set up venv and requirements
if [ -d "$SCRIPT_DIR/venv" ]; then
  echo "using pre-existing venv for keycloak ais-admin creation script"
  source "$SCRIPT_DIR/venv/bin/activate"
else
  echo "venv not found, creating and installing requirements for keycloak ais-admin creation script"
  python3 -m venv "$SCRIPT_DIR/venv"
  source "$SCRIPT_DIR/venv/bin/activate"
  pip install python-keycloak
fi

# Build python arguments, conditionally add --verify-ca
PY_ARGS=(
  "$SCRIPT_DIR/create_ais_admin.py"
  --host "$KEYCLOAK_HOST"
  --realm aistore
  --admin-user "$USER"
)

if [ -n "$CA_FILE" ]; then
  PY_ARGS+=( --verify-ca "$CA_FILE" )
fi

python "${PY_ARGS[@]}"
