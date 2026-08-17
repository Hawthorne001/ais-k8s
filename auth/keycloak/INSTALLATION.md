# Installation steps for Keycloak on K8s

Use `./test-cluster.sh` to automatically provision a local KinD cluster with Keycloak installed and ready for AIS.

The manual steps required are documented below.

## Prerequisites

### External Access

The sample manifest disables ingress, and this guide reaches Keycloak through the cluster-internal service.
To expose it outside the cluster instead, configure an [ingress controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) and set `spec.ingress.enabled: true`.

### Cert-manager

For deploying with TLS with the provided certificate, you'll need [cert-manager](https://cert-manager.io/) with an `Issuer` or `ClusterIssuer` configured. 

### Database

Keycloak requires a relational DB to persist data and share between instances. 
A list of compatible DBs can be found [here](https://www.keycloak.org/server/db).

For this reference installation, follow the [CloudNativePG](https://cloudnative-pg.io/documentation/current/quickstart/) quickstart guide.
We have a sample helmfile for the CNPG operator and cluster as well as some default values in [helm/cnpg/](./helm/cnpg/).

This default installation includes an `app` database and user, which we will use for Keycloak.
If needed, use the superuser and its associated secret to login and create a new user for keycloak.
Get the password for the `app` user created in a secret by the CNPG installation:


    kubectl get secrets cloudnative-pg-cluster-app --namespace cnpg-database --template={{.data.password}} | base64 -d


## Keycloak Installation 

1. Install the [Keycloak operator](https://www.keycloak.org/operator/installation). 
1. Create a secret for Keycloak to use for DB access
    1. Shown here with the CNPG default username and password from secret: 
    ```console
    kubectl create secret -n keycloak generic keycloak-db-secret --from-literal=username=app --from-literal=password="$(kubectl get secret cloudnative-pg-cluster-app --namespace cnpg-database -o jsonpath='{.data.password}' | base64 --decode)"
    ``` 
1. Create TLS certificate to use
    1. [Sample certificate manifest](./manifests/certificate.yaml).
1. Install Keycloak 
    1. Check [Keycloak's documentation](https://www.keycloak.org/operator/basic-deployment) for manifest options
    1. [Sample manifest](./manifests/keycloak.yaml)
1. Log in. 
    1. Forward the service to your local machine
    1.  `kubectl port-forward -n keycloak service/keycloak-server-service 8543:8543`
    1. If you run kubectl from a remote host rather than your local machine, SSH tunnel to that host
    1.  `ssh -L 8543:127.0.0.1:8543 <your node hostname>`
    1. API requests work directly against `https://localhost:8543`
    1. For browser access, you'll need an etc/hosts entry: `127.0.0.1 <keycloak hostname>`
    1. Go to this address in your browser `https://<keycloak hostname>:8543/`
    1. Get the admin login info from k8s
    ```
    kubectl get secret -n keycloak keycloak-server-initial-admin -o jsonpath='{.data.username}' | base64 --decode
    kubectl get secret -n keycloak keycloak-server-initial-admin -o jsonpath='{.data.password}' | base64 --decode
    ```
        > Note: If you re-used a db that already has this user, the k8s secrets may not contain the password.
1. Import the aistore prebuilt realm
    1. We provide a script to follow the [Keycloak docs](https://www.keycloak.org/operator/realm-import) and automate importing from any exported realm JSON file. 
    1. Run `./realm/import-realm.sh` to import a default **development** AIS realm.
    1. If desired, run `./scripts/prepare_cluster.sh` to provision a default admin user. See the [main readme](./README.md#aistore-realm) for details.  