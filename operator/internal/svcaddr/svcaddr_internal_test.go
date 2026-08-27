/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

package svcaddr

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

const testDomain = "k8s.example.com"

var testSvc = types.NamespacedName{Name: "ais-proxy", Namespace: "ais-ns"}

func stubClusterDomain(domain string) {
	previous := clusterDomain
	DeferCleanup(func() { clusterDomain = previous })
	clusterDomain = func() string { return domain }
}

var _ = Describe("Service addresses", Label("short"), func() {
	BeforeEach(func() {
		stubClusterDomain(testDomain)
	})

	It("should qualify a Service with the cluster domain", func() {
		Expect(ServiceFQDN(testSvc)).To(Equal("ais-proxy.ais-ns.svc." + testDomain))
	})

	It("should qualify every Pod behind a headless Service", func() {
		Expect(WildcardServiceFQDN(testSvc)).To(Equal("*.ais-proxy.ais-ns.svc." + testDomain))
	})

	It("should address a Service on its port", func() {
		Expect(ServiceURL("https", testSvc, "51080")).To(Equal("https://ais-proxy.ais-ns.svc." + testDomain + ":51080"))
	})

	It("should address a Pod through its headless Service", func() {
		Expect(PodURL("http", "ais-proxy-0", testSvc, "51082")).
			To(Equal("http://ais-proxy-0.ais-proxy.ais-ns.svc." + testDomain + ":51082"))
	})
})
