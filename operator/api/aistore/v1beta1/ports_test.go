/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

package v1beta1

import (
	aisapc "github.com/NVIDIA/aistore/api/apc"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var _ = Describe("Ports", func() {
	Describe("Defaults", func() {
		It("should default every port when the spec omits them", func() {
			ais := AIStore{}
			Expect(ais.ProxyPublicPort()).To(Equal(intstr.FromInt32(51080)))
			Expect(ais.TargetPublicPort()).To(Equal(intstr.FromInt32(51081)))
			Expect(ais.ProxyIntraControlPort()).To(Equal(intstr.FromInt32(51082)))
			Expect(ais.TargetIntraControlPort()).To(Equal(intstr.FromInt32(51082)))
			Expect(ais.ProxyIntraDataPort()).To(Equal(intstr.FromInt32(51083)))
			Expect(ais.TargetIntraDataPort()).To(Equal(intstr.FromInt32(51083)))
		})

		It("should keep ports set by the spec", func() {
			ais := AIStore{Spec: AIStoreSpec{
				ProxySpec: DaemonSpec{ServiceSpec: ServiceSpec{
					PublicPort:       aisapc.Ptr(intstr.FromInt32(9080)),
					IntraControlPort: aisapc.Ptr(intstr.FromInt32(9082)),
				}},
			}}
			Expect(ais.ProxyPublicPort()).To(Equal(intstr.FromInt32(9080)))
			Expect(ais.ProxyIntraControlPort()).To(Equal(intstr.FromInt32(9082)))
			Expect(ais.ProxyIntraDataPort()).To(Equal(intstr.FromInt32(51083)))
		})

		It("should not default a port explicitly set to zero", func() {
			ais := AIStore{Spec: AIStoreSpec{
				ProxySpec: DaemonSpec{ServiceSpec: ServiceSpec{
					PublicPort: aisapc.Ptr(intstr.FromInt32(0)),
				}},
			}}
			Expect(ais.ProxyPublicPort()).To(Equal(intstr.FromInt32(0)))
		})
	})
})
