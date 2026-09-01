/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

package cmn

import (
	"fmt"
	"net"

	aisv1 "github.com/ais-operator/api/aistore/v1beta1"
	"github.com/ais-operator/internal/svcaddr"
)

// DefaultProxyURL returns the URL of the proxy that starts out as primary.
func DefaultProxyURL(ais *aisv1.AIStore) string {
	// Example: https://ais-proxy-0.ais-proxy.ais.svc.cluster.local:51082
	controlPort := ais.ProxyIntraControlPort()
	return svcaddr.PodURL(urlScheme(ais), ais.DefaultPrimaryName(),
		ais.ProxyHeadlessSVCNSName(), controlPort.String())
}

// IntraClusterURL returns the URL of the cluster-internal proxy service on the public network.
func IntraClusterURL(ais *aisv1.AIStore) string {
	// Example: https://ais-proxy.ais.svc.cluster.local:51080
	publicPort := ais.ProxyPublicPort()
	return svcaddr.ServiceURL(urlScheme(ais),
		ais.ProxyHeadlessSVCNSName(), publicPort.String())
}

// DiscoveryProxyURL returns the URL of the proxy service on the intra-control network.
func DiscoveryProxyURL(ais *aisv1.AIStore) string {
	// Example: https://ais-proxy.ais.svc.cluster.local:51082
	controlPort := ais.ProxyIntraControlPort()
	return svcaddr.ServiceURL(urlScheme(ais),
		ais.ProxyHeadlessSVCNSName(), controlPort.String())
}

// AISHostURL returns the URL matching the AIS scheme with the given hostname and port.
// The hostname may be a DNS name or an IPv4 or IPv6 literal.
func AISHostURL(ais *aisv1.AIStore, hostname, port string) string {
	return fmt.Sprintf("%s://%s", urlScheme(ais), net.JoinHostPort(hostname, port))
}

func urlScheme(ais *aisv1.AIStore) string {
	if ais.UseHTTPS() {
		return "https"
	}
	return "http"
}
