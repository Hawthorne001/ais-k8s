/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

package cmn

import (
	aisv1 "github.com/ais-operator/api/aistore/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Env", Label("short"), func() {
	Describe("CommonEnv", func() {
		It("returns the node, pod, namespace and cluster name vars", func() {
			ais := &aisv1.AIStore{ObjectMeta: metav1.ObjectMeta{Name: "test-cluster", Namespace: "ais"}}

			Expect(CommonEnv(ais)).To(Equal([]corev1.EnvVar{
				EnvFromFieldPath(EnvNodeName, "spec.nodeName"),
				EnvFromFieldPath(EnvPodName, "metadata.name"),
				EnvFromFieldPath(EnvNS, "metadata.namespace"),
				EnvFromValue(EnvCluster, "test-cluster"),
			}))
		})

		It("sources the cluster name from the resource name", func() {
			ais := &aisv1.AIStore{ObjectMeta: metav1.ObjectMeta{Name: "cluster-name", Namespace: "ns-name"}}
			env := CommonEnv(ais)
			Expect(env).To(ContainElement(corev1.EnvVar{Name: EnvCluster, Value: "cluster-name"}))
		})
	})
	Describe("CommonInitEnv", func() {
		It("includes the cluster name var", func() {
			ais := &aisv1.AIStore{ObjectMeta: metav1.ObjectMeta{Name: "cluster-name"}}
			env := CommonInitEnv(ais, false)
			Expect(env).To(ContainElement(corev1.EnvVar{Name: EnvCluster, Value: "cluster-name"}))
		})

		It("includes the host IPs var", func() {
			Expect(CommonInitEnv(&aisv1.AIStore{}, false)).To(
				ContainElement(EnvFromFieldPath(EnvHostIPS, "status.hostIPs")))
		})

		It("sets the public DNS mode from the spec", func() {
			mode := aisv1.PubNetDNSModeNode
			ais := &aisv1.AIStore{Spec: aisv1.AIStoreSpec{PublicNetDNSMode: &mode}}

			Expect(CommonInitEnv(ais, false)).To(ContainElement(EnvFromValue(EnvPublicDNSMode, string(mode))))
		})

		It("omits the public DNS mode when unset", func() {
			Expect(CommonInitEnv(&aisv1.AIStore{}, false)).NotTo(
				ContainElement(HaveField("Name", EnvPublicDNSMode)))
		})
	})

})
