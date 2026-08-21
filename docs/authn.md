# AIStore Authentication Server (AuthN) in Kubernetes

>  **NOTE**: AuthN and its related deployment automations are under development. Breaking changes are to be expected, and it has NOT gone through a complete security audit.
Please review your deployment carefully and follow [our security policy](https://github.com/NVIDIA/ais-k8s/blob/main/SECURITY.md) to report any issues.

The AIStore Authentication Server (AuthN) provides secure access to AIStore by leveraging [OAuth 2.0](https://oauth.net/2/) compliant [JSON Web Tokens (JWT)](https://datatracker.ietf.org/doc/html/rfc7519). 

For more information on AuthN, visit the [AIStore AuthN documentation](https://github.com/NVIDIA/aistore/blob/main/docs/authn.md).

## Deploying AuthN in Kubernetes

### Deploy with Helm

The best way to deploy authN is to use our [provided Helm chart](../helm/authn/README.md)

### AuthN Resources in Kubernetes

- **Static Resources**
  - **Signing Key Secret**  
     - This secret holds the key used to sign JWT tokens, which is used by both the AuthN server and AIStore pods.
  - **Admin Credentials Secret**
     - This secret contains the admin user and password as entries, mapped to `SU-NAME` and `SU-PASS`.
  - **AuthN Configuration ConfigMap**  
     - The ConfigMap stores the non-sensitive default configuration of the AuthN server.
  - **Persistent Storage (PV and PVC)**  
     - User information and configuration data for AuthN are stored in a Persistent Volume (PV), which is connected to the AuthN deployment via a Persistent Volume Claim (PVC).
- **Services**
  - **External Service for AuthN**
    - This service exposes the AuthN server to external clients. You can choose to use either a `NodePort` or `LoadBalancer` service, depending on your access requirements.
  - **Internal Service for AuthN**
     - This service facilitates internal communication between the AuthN server and other pods, including the AIS-Operator, within the cluster.
- **AuthN Deployment**  
   - This runs the AuthN pod and connects it with the other resources.

## AuthN Clients

See the [authentication docs](./authentication.md) for information about configuring the AIStore cluster and operator to use authN and other authentication services.

To interact with AIStore, clients need a signed JWT token.
By default, an `admin` user with super-user privileges is created with a mandatory provided password.
This password must be set through [environment variables](https://github.com/NVIDIA/aistore/blob/main/docs/authn.md#environment-and-configuration).
Admins can then create roles and assign users to those roles.
For a typical setup process, refer to the [Getting Started Guide](https://github.com/NVIDIA/aistore/blob/main/docs/authn.md#getting-started).

Set the following environment variable to point to the appropriate AuthN server to log in and obtain the token:

```bash
# For external clients
export AIS_AUTHN_URL=https://<NodePort-service-IP>:30001

# For internal clients
export AIS_AUTHN_URL=https://ais-authn.ais:52001
```

## Switching Between HTTP and HTTPS (TLS) for the AuthN Server

For how AuthN certificates are issued and trusted, see the [TLS guide](./tls.md).

To switch the protocol of an existing AuthN server from HTTP to HTTPS (or vice versa), you can apply the new configuration specification over the current deployment.
This will automatically redeploy the AuthN server with the updated settings.

We recommend using the [AuthN Helm chart](../helm/authn/README.md) for this process.

This will require an update to `spec.serviceURL` and potentially `spec.tls` in the `AIStoreAuthProfile` referenced by the AIStore spec.
