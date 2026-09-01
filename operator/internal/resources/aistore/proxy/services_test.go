/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

package proxy

import (
	aisapc "github.com/NVIDIA/aistore/api/apc"
	aisv1 "github.com/ais-operator/api/aistore/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("Proxy NewProxyLoadBalancerSVC", func() {
	newAIS := func(proxy aisv1.DaemonSpec) *aisv1.AIStore {
		return &aisv1.AIStore{
			ObjectMeta: metav1.ObjectMeta{Name: "test-ais", Namespace: "test-ns"},
			Spec:       aisv1.AIStoreSpec{ProxySpec: proxy},
		}
	}

	It("should publish the LoadBalancer port while targeting the public port", func() {
		ais := newAIS(aisv1.DaemonSpec{
			ServiceSpec: aisv1.ServiceSpec{PublicPort: aisapc.Ptr(intstr.FromInt32(9080))},
			ExternalAccess: &aisv1.ExternalAccessSpec{
				LoadBalancer: &aisv1.LoadBalancerSpec{Port: aisapc.Ptr[int32](443)},
			},
		})

		ports := NewProxyLoadBalancerSVC(ais).Spec.Ports
		Expect(ports).To(HaveLen(1))
		Expect(*ports[0].Port).To(Equal(int32(443)))
		Expect(*ports[0].TargetPort).To(Equal(intstr.FromInt32(9080)))
	})

	It("should publish the public port when the LoadBalancer port is unset", func() {
		ais := newAIS(aisv1.DaemonSpec{ExternalAccess: &aisv1.ExternalAccessSpec{}})

		ports := NewProxyLoadBalancerSVC(ais).Spec.Ports
		Expect(ports).To(HaveLen(1))
		Expect(*ports[0].Port).To(Equal(aisv1.DefaultProxyPublicPort))
		Expect(*ports[0].TargetPort).To(Equal(intstr.FromInt32(aisv1.DefaultProxyPublicPort)))
	})
})
