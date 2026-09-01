/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

package v1beta1

import "k8s.io/apimachinery/pkg/util/intstr"

// Container ports the AIS proxy and target processes listen on when the spec omits them.
const (
	DefaultProxyPublicPort  int32 = 51080
	DefaultTargetPublicPort int32 = 51081
	DefaultIntraControlPort int32 = 51082
	DefaultIntraDataPort    int32 = 51083
)

func portOrDefault(port *intstr.IntOrString, def int32) intstr.IntOrString {
	if port == nil {
		return intstr.FromInt32(def)
	}
	return *port
}

// daemonPorts holds one daemon's ports with the defaults already applied.
type daemonPorts struct {
	service      intstr.IntOrString
	public       intstr.IntOrString
	intraControl intstr.IntOrString
	intraData    intstr.IntOrString
}

func (ais *AIStore) proxyPorts() *daemonPorts {
	return &daemonPorts{
		service:      ais.Spec.ProxySpec.ServicePort,
		public:       ais.ProxyPublicPort(),
		intraControl: ais.ProxyIntraControlPort(),
		intraData:    ais.ProxyIntraDataPort(),
	}
}

func (ais *AIStore) targetPorts() *daemonPorts {
	return &daemonPorts{
		service:      ais.Spec.TargetSpec.ServicePort,
		public:       ais.TargetPublicPort(),
		intraControl: ais.TargetIntraControlPort(),
		intraData:    ais.TargetIntraDataPort(),
	}
}

// ProxyPublicPort returns the container port the proxy process listens on for the public network.
func (ais *AIStore) ProxyPublicPort() intstr.IntOrString {
	return portOrDefault(ais.Spec.ProxySpec.PublicPort, DefaultProxyPublicPort)
}

// ProxyIntraControlPort returns the container port the proxy process listens on for the
// intra-cluster control network.
func (ais *AIStore) ProxyIntraControlPort() intstr.IntOrString {
	return portOrDefault(ais.Spec.ProxySpec.IntraControlPort, DefaultIntraControlPort)
}

// ProxyIntraDataPort returns the container port the proxy process listens on for the
// intra-cluster data network.
func (ais *AIStore) ProxyIntraDataPort() intstr.IntOrString {
	return portOrDefault(ais.Spec.ProxySpec.IntraDataPort, DefaultIntraDataPort)
}

// TargetPublicPort returns the container port the target process listens on for the public network.
func (ais *AIStore) TargetPublicPort() intstr.IntOrString {
	return portOrDefault(ais.Spec.TargetSpec.PublicPort, DefaultTargetPublicPort)
}

// TargetIntraControlPort returns the container port the target process listens on for the
// intra-cluster control network.
func (ais *AIStore) TargetIntraControlPort() intstr.IntOrString {
	return portOrDefault(ais.Spec.TargetSpec.IntraControlPort, DefaultIntraControlPort)
}

// TargetIntraDataPort returns the container port the target process listens on for the
// intra-cluster data network.
func (ais *AIStore) TargetIntraDataPort() intstr.IntOrString {
	return portOrDefault(ais.Spec.TargetSpec.IntraDataPort, DefaultIntraDataPort)
}
