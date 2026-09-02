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
	service      *intstr.IntOrString
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

func externalPortOrDefault(spec *DaemonSpec, publicPort intstr.IntOrString) intstr.IntOrString {
	if ea := spec.ExternalAccess; ea != nil && ea.LoadBalancer != nil && ea.LoadBalancer.Port != nil {
		return intstr.FromInt32(*ea.LoadBalancer.Port)
	}
	// Keep pre-existing LoadBalancers on the port servicePort published.
	if spec.ServicePort != nil {
		return *spec.ServicePort
	}
	return publicPort
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

// ProxyExternalPort returns the port the proxy LoadBalancer service listens on.
func (ais *AIStore) ProxyExternalPort() intstr.IntOrString {
	return externalPortOrDefault(&ais.Spec.ProxySpec, ais.ProxyPublicPort())
}

// TargetExternalPort returns the port the target LoadBalancer services listen on.
func (ais *AIStore) TargetExternalPort() intstr.IntOrString {
	return externalPortOrDefault(&ais.Spec.TargetSpec.DaemonSpec, ais.TargetPublicPort())
}

// DeprecatedPortMessages returns a message for each deprecated port option set.
func (s *AIStoreSpec) DeprecatedPortMessages() []string {
	var msgs []string
	if s.ProxySpec.ServicePort != nil {
		msgs = append(msgs, "spec.proxySpec.servicePort is deprecated, use spec.proxySpec.portPublic "+
			"and spec.proxySpec.externalAccess.loadBalancer.port")
	}
	if s.TargetSpec.ServicePort != nil {
		msgs = append(msgs, "spec.targetSpec.servicePort is deprecated, use spec.targetSpec.portPublic")
	}
	return msgs
}
