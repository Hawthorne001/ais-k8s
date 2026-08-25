# Kube State Metrics

## Overview

[Kube State Metrics](https://github.com/kubernetes/kube-state-metrics) (KSM) exposes Prometheus metrics about the state of Kubernetes API objects (e.g. Deployments, Pods, Nodes). This Helmfile deploys KSM as a standalone component in the `monitoring` namespace for Alloy to scrape.

The [Custom Resource State](https://github.com/kubernetes/kube-state-metrics/blob/main/docs/metrics/extend/customresourcestate-metrics.md) config in `values.yaml.gotmpl` additionally exports the AIStore CR's `status.conditions` and `status.state` as `kube_customresource_aistore_*`. Environments that set a `metricAllowlist` must list those names for them to reach Prometheus.

## Usage

Template manifests:

```bash
helmfile -e prod template
```

Deploy/sync:

```bash
helmfile -e prod sync
```

Port-forward for quick inspection (optional):

```bash
kubectl -n monitoring port-forward svc/kube-state-metrics 8080:8080
curl -s localhost:8080/metrics | head
```

# Relevant Links

- [Chart source](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-state-metrics)
- [Default values](https://github.com/prometheus-community/helm-charts/blob/main/charts/kube-state-metrics/values.yaml)
