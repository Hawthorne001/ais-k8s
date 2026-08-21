# Deploying with Authentication

This doc covers deployment of an AIStore cluster with authentication enabled in K8s. 

Deployment of the token issuer is managed separately.
This may be [AuthN](./authn.md), [Keycloak](./deploy_with_keycloak.md), or any other issuer configured to issue [AIS-compatible JWTs](https://github.com/NVIDIA/aistore/blob/main/docs/auth_validation.md).

## Enabling AIStore Authentication

See the [AIS authentication docs](https://github.com/NVIDIA/aistore/blob/main/docs/auth_validation.md#authentication-flow) for information about how to provision and validate tokens.
See the [v5.0 release notes](https://github.com/NVIDIA/aistore/releases/tag/v1.5.0#intra-cluster-security) for more information about intra-cluster security changes included with AIStore v5.0.

### Public Key Signature Verification

For [RSA public key signature verification](https://github.com/NVIDIA/aistore/blob/main/docs/auth_validation.md#static-credentials) or [OIDC issuer lookup](https://github.com/NVIDIA/aistore/blob/main/docs/auth_validation.md#oidc-lookup), set the values directly in the `spec.configToUpdate.auth` section of the AIStore resource. 

#### Static RSA Public Key

```yaml
spec:
  configToUpdate:
    auth: 
      enabled: true
      # Static JWT signature configuration -- mutually exclusive with "oidc" option
      signature:
        method: RSA
        key: <RSA Pub Key if using static key pair>
```

#### OIDC Issuer Lookup

```yaml
spec:
  configToUpdate:
    auth: 
      enabled: true
      # Dynamic JWKS lookup for JWT signature validation
      oidc:
        allowed_iss: ["<accessible issuer OIDC discovery URL>"]
```

### Symmetric Key Signature Verification

If using HMAC signing, create a secret with the key `SIGNING-KEY` in the AIStore namespace.

Set `spec.authNSecretName` to reference this secret, and the operator will set the `AIS_AUTHN_SECRET_KEY` environment variable in the AIStore proxy containers to the referenced value.

```yaml
spec:
  authNSecretName: jwt-signing-key
  configToUpdate:
    auth: 
      enabled: true
      signature: 
        method: HMAC
```

## AIS Operator Access

If authentication is enabled for your AIStore cluster, AIS Operator requires an admin token since it frequently calls AIStore lifecycle APIs.

### AIStoreAuthProfile

The recommended way to configure this for each managed AIS cluster is to create and reference an `AIStoreAuthProfile` custom resource.
See the [AIStoreAuthProfile guide](./auth_profile.md) for details and examples. 

Once created, the profile can be easily referenced and used by anyone with `use` RBAC access to that profile: 

```yaml
spec:
  auth:
    profileRef:
      name: local-admin
```

### Legacy `spec.auth` fields (deprecated)

> **NOTE:** `spec.auth` fields described below are now **deprecated**.
> See the [AIStoreAuthProfile profile guide](./auth_profile.md) and configure `spec.auth.profileRef` instead.

Legacy versions of the AIStore custom resource supported configuring the Operator's authentication directly in spec. 
These options are deprecated but documented below for reference. 

#### Username/Password Authentication

Specify the location of the admin credentials secret directly in the AIS spec for each cluster.
For examples of `auth.usernamePassword` see the auth section in the [provided config examples](../operator/config/samples/aistore_auth_legacy.yaml).

#### Token Exchange Authentication

Exchanging a token with the authentication service for an AIS JWT token eliminates the need to store static admin credentials.
This mode requires the authentication service to support a token exchange endpoint (default: `/token`).

On versions 3.3 and below, this token is read from the filesystem (e.g., Kubernetes service account token or OIDC token).
Later versions request a short-lived token for the operator's own ServiceAccount.

For configuring token exchange in the AIS spec see `auth.tokenExchange` in the [provided config examples](../operator/config/samples/aistore_auth_legacy.yaml)

**Mounting Custom Tokens:**
To use a custom OIDC token, add a projected volume to the operator deployment:
```yaml
volumes:
- name: oidc-token
  projected:
    sources:
    - serviceAccountToken:
        path: token
        expirationSeconds: 3600
        audience: ais-authn
```

`tokenPath` is empty by default, and the operator requests a short-lived token for its own ServiceAccount. 
For versions 3.3 and below, it defaulted to `/var/run/secrets/kubernetes.io/serviceaccount/token`.
Set `tokenPath` to the projected path to tell the operator to reference it instead of minting a new token.

## Disabling Authentication in an Existing AIStore Deployment

If you have authentication enabled but no longer wish to use it:
1. Disable the config option in spec by setting `spec.configToUpdate.auth.enabled: false`.
1. Remove any referenced `spec.authNSecretName`.
1. Remove any remaining config under `spec.auth`.

## Enabling Auth on a Running AIStore Server

> Note: If using `authNSecretName` for HMAC signing, adding the secret will cause a rollout of all proxy pods

1. Deploy your token issuer.
   1. If deploying AIS AuthN, see the [Helm chart](../helm/authn/README.md).
1. Create an [AIStoreAuthProfile](#aistoreauthprofile) for the deployment.
1. Update the AIS custom resource 
   1. Set `spec.auth.profileRef` to reference the new `AIStoreAuthProfile`.
   1. Set `spec.configToUpdate.auth.enabled: false` to prevent enabling AIStore-side authentication until the signature validation is configured.  
1. [Enable authentication validation](#enabling-aistore-authentication).
