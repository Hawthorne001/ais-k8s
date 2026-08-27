/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

// Package svcaddr builds the addresses under which in-cluster Services and their Pods are reachable.
package svcaddr

import (
	"fmt"

	"github.com/ais-operator/internal/opinfo"
	"k8s.io/apimachinery/pkg/types"
)

// tests substitute their own domain
var clusterDomain = opinfo.ClusterDomain

// ServiceFQDN returns the fully qualified DNS name of a Service.
func ServiceFQDN(svc types.NamespacedName) string {
	return fmt.Sprintf("%s.%s.svc.%s", svc.Name, svc.Namespace, clusterDomain())
}

// WildcardServiceFQDN returns the DNS name matching every Pod behind a headless Service.
func WildcardServiceFQDN(svc types.NamespacedName) string {
	return "*." + ServiceFQDN(svc)
}

// ServiceURL returns the URL a Service answers on within the cluster.
func ServiceURL(scheme string, svc types.NamespacedName, port string) string {
	return fmt.Sprintf("%s://%s:%s", scheme, ServiceFQDN(svc), port)
}

// PodURL returns the URL a Pod answers on within the cluster, addressed through its headless Service.
func PodURL(scheme, pod string, svc types.NamespacedName, port string) string {
	return fmt.Sprintf("%s://%s.%s:%s", scheme, pod, ServiceFQDN(svc), port)
}
