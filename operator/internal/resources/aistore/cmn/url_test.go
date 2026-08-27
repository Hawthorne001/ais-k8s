/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

package cmn_test

import (
	"fmt"

	aisapc "github.com/NVIDIA/aistore/api/apc"
	aisv1 "github.com/ais-operator/api/aistore/v1beta1"
	"github.com/ais-operator/internal/resources/aistore/cmn"
	"github.com/ais-operator/internal/svcaddr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	urlTestPublicPort   = 51080
	urlTestControlPort  = 51082
	urlTestScheme       = "http"
	urlTestSecureScheme = "https"
	urlTestName         = "ais"
	urlTestNamespace    = "ais-ns"
	urlTestPrimaryPod   = "ais-proxy-0"
)

var urlTestProxySvc = types.NamespacedName{Name: "ais-proxy", Namespace: urlTestNamespace}

func newURLTestAIS() *aisv1.AIStore {
	return &aisv1.AIStore{
		ObjectMeta: metav1.ObjectMeta{Name: urlTestName, Namespace: urlTestNamespace},
		Spec: aisv1.AIStoreSpec{
			ProxySpec: aisv1.DaemonSpec{
				ServiceSpec: aisv1.ServiceSpec{
					PublicPort:       intstr.FromInt32(urlTestPublicPort),
					IntraControlPort: intstr.FromInt32(urlTestControlPort),
				},
			},
		},
	}
}

var _ = Describe("Proxy URLs", Label("short"), func() {
	var ais *aisv1.AIStore

	BeforeEach(func() {
		ais = newURLTestAIS()
	})

	It("should address the primary proxy pod on the intra-control port", func() {
		Expect(cmn.DefaultProxyURL(ais)).To(Equal(fmt.Sprintf("%s://%s.%s:%d",
			urlTestScheme, urlTestPrimaryPod, svcaddr.ServiceFQDN(urlTestProxySvc), urlTestControlPort)))
	})

	It("should address the proxy service on the intra-control port", func() {
		Expect(cmn.DiscoveryProxyURL(ais)).To(Equal(fmt.Sprintf("%s://%s:%d",
			urlTestScheme, svcaddr.ServiceFQDN(urlTestProxySvc), urlTestControlPort)))
	})

	It("should address the proxy service on the public port", func() {
		Expect(cmn.IntraClusterURL(ais)).To(Equal(fmt.Sprintf("%s://%s:%d",
			urlTestScheme, svcaddr.ServiceFQDN(urlTestProxySvc), urlTestPublicPort)))
	})

	It("should use https when the cluster does", func() {
		ais.Spec.ConfigToUpdate = &aisv1.ConfigToUpdate{
			Net: &aisv1.NetConfToUpdate{
				HTTP: &aisv1.HTTPConfToUpdate{UseHTTPS: aisapc.Ptr(true)},
			},
		}

		Expect(cmn.IntraClusterURL(ais)).To(HavePrefix(urlTestSecureScheme + "://"))
	})
})
