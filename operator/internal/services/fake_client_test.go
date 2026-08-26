/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

package services

import (
	authv1alpha1 "github.com/ais-operator/api/aisauth/v1alpha1"
	aisclient "github.com/ais-operator/internal/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
)

func NewFakeK8sClient(objs ...client.Object) *aisclient.K8sClient {
	GinkgoHelper()
	return NewFakeK8sClientWithInterceptors(nil, objs...)
}

func NewFakeK8sClientWithInterceptors(funcs *interceptor.Funcs, objs ...client.Object) *aisclient.K8sClient {
	GinkgoHelper()
	sch := runtime.NewScheme()
	Expect(scheme.AddToScheme(sch)).To(Succeed())
	Expect(authv1alpha1.AddToScheme(sch)).To(Succeed())
	builder := fake.NewClientBuilder().WithScheme(sch).WithObjects(objs...)
	if funcs != nil {
		builder = builder.WithInterceptorFuncs(*funcs)
	}
	return aisclient.NewClient(builder.Build(), sch)
}
