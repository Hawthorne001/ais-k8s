// Package e2e contains AIS operator integration tests
/*
 * Copyright (c) 2025, NVIDIA CORPORATION. All rights reserved.
 */
package e2e

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	aisapi "github.com/NVIDIA/aistore/api"
	aisapc "github.com/NVIDIA/aistore/api/apc"
	aistutils "github.com/NVIDIA/aistore/tools"
	aisv1 "github.com/ais-operator/api/v1beta1"
	aisclient "github.com/ais-operator/pkg/client"
	"github.com/ais-operator/pkg/resources/proxy"
	"github.com/ais-operator/pkg/resources/target"
	"github.com/ais-operator/tests/tutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clientpkg "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	clusterCreateInterval     = time.Second
	clusterReadyRetryInterval = 5 * time.Second
	clusterReadyTimeout       = 5 * time.Minute
	clusterDestroyInterval    = 2 * time.Second
	clusterDestroyTimeout     = 3 * time.Minute
	clusterUpdateTimeout      = 1 * time.Minute
	clusterUpdateInterval     = 2 * time.Second

	urlTemplate = "http://%s:%s"
)

// clientCluster - This struct contains an AIS custom resource, references to required persistent volumes,
// and utility methods for managing clusters used by operator tests
type clientCluster struct {
	aisCtx    *tutils.AISTestContext
	k8sClient *aisclient.K8sClient
	cluster   *aisv1.AIStore
	pvs       []*corev1.PersistentVolume
	ctx       context.Context
	proxyURL  string
}

func (cc *clientCluster) applyDefaultHostPortOffset(args *tutils.ClusterSpecArgs) {
	if args.EnableExternalLB {
		return
	}
	// Apply host port offset of 10 per parallel Ginkgo process to give each process a unique host port
	// and allow for further in-test offsets (e.g. multiple clusters in the same test)
	gid := int32(GinkgoParallelProcess())
	cc.applyHostPortOffset(gid * 10)
}

func newClientCluster(ctx context.Context, aisCtx *tutils.AISTestContext, k8sClient *aisclient.K8sClient, cluArgs *tutils.ClusterSpecArgs) *clientCluster {
	cluster, pvs := tutils.NewAISCluster(cluArgs, k8sClient)
	cc := &clientCluster{
		ctx:       ctx,
		aisCtx:    aisCtx,
		k8sClient: k8sClient,
		cluster:   cluster,
		pvs:       pvs,
	}
	cc.applyDefaultHostPortOffset(cluArgs)
	return cc
}

func (cc *clientCluster) getTimeout() time.Duration {
	// For a cluster with external LB, allocating external-IP could be time-consuming.
	// Force longer timeout for cluster creation.
	if cc.cluster.Spec.EnableExternalLB {
		return cc.aisCtx.GetClusterCreateLongTimeout()
	}
	return cc.aisCtx.GetClusterCreateTimeout()
}

// Use to avoid a host port collision with an existing host port cluster
func (cc *clientCluster) applyHostPortOffset(offset int32) {
	specs := []*aisv1.DaemonSpec{&cc.cluster.Spec.ProxySpec, &cc.cluster.Spec.TargetSpec.DaemonSpec}
	for i := range specs {
		specs[i].HostPort = aisapc.Ptr(*specs[i].HostPort + offset)
		specs[i].ServicePort = intstr.FromInt32(specs[i].ServicePort.IntVal + offset)
		specs[i].PublicPort = intstr.FromInt32(specs[i].PublicPort.IntVal + offset)
	}
}

// Re-initialize the local cluster CR from the given cluster args and re-create it remotely -- does not create PVs
func (cc *clientCluster) recreate(cluArgs *tutils.ClusterSpecArgs) {
	cc.cluster = tutils.NewAISClusterNoPV(cluArgs)
	cc.applyDefaultHostPortOffset(cluArgs)
	cc.create()
}

func (cc *clientCluster) create() {
	cc.createCluster(cc.getTimeout(), clusterCreateInterval)
	cc.waitForReadyCluster()
	cc.initClientAccess()
}

func createClusters(clusters []*clientCluster) {
	var wg sync.WaitGroup
	wg.Add(len(clusters))

	for _, cluster := range clusters {
		go func(cc *clientCluster) {
			defer GinkgoRecover()
			defer wg.Done()
			cc.create()
		}(cluster)
	}
	wg.Wait()
}

func (cc *clientCluster) createWithCallback(postCreate func()) {
	cc.create()
	if postCreate != nil {
		By("Running post-create callback")
		postCreate()
	}
}

func (cc *clientCluster) createAndDestroyCluster(postCreate, postDestroy func()) {
	defer func() {
		Expect(cc.printLogs()).To(Succeed())
		cc.destroyCleanupWithCallback(postDestroy)
	}()
	cc.createWithCallback(postCreate)
}

func (cc *clientCluster) createCluster(intervals ...interface{}) {
	Expect(cc.k8sClient.Create(cc.ctx, cc.cluster)).Should(Succeed())
	By("Create cluster and wait for it to be 'Ready'")
	Eventually(func() bool {
		ais := &aisv1.AIStore{}
		_ = cc.k8sClient.Get(cc.ctx, cc.cluster.NamespacedName(), ais)
		return ais.HasState(aisv1.ClusterReady)
	}, intervals...).Should(BeTrue())
}

func (cc *clientCluster) refresh() {
	var err error
	cc.cluster, err = cc.k8sClient.GetAIStoreCR(cc.ctx, cc.cluster.NamespacedName())
	Expect(err).NotTo(HaveOccurred())
}

func (cc *clientCluster) waitForReadyCluster() {
	tutils.WaitForClusterToBeReady(cc.ctx, cc.k8sClient, cc.cluster.NamespacedName(), clusterReadyTimeout, clusterReadyRetryInterval)
	// Validate the cluster map -- make sure all AIS nodes have successfully joined cluster
	cc.refresh()
	cc.initClientAccess()

	proxyURLs := cc.getAllProxyURLs()
	expectedProxies := cc.cluster.GetProxySize()
	expectedTargets := cc.cluster.GetTargetSize()

	for i := range proxyURLs {
		proxyURL := *proxyURLs[i]
		bp := aistutils.BaseAPIParams(proxyURL)
		Eventually(func() bool {
			smap, err := aisapi.GetClusterMap(bp)
			if err != nil {
				return false
			}
			activeProxies := int32(len(smap.Pmap.ActiveNodes()))
			activeTargets := int32(len(smap.Tmap.ActiveNodes()))
			return activeProxies == expectedProxies && activeTargets == expectedTargets
		}).Should(BeTrue())
	}
}

func (cc *clientCluster) patchImagesToCurrent() {
	cc.fetchLatestCluster()
	patch := clientpkg.MergeFrom(cc.cluster.DeepCopy())
	cc.cluster.Spec.NodeImage = cc.aisCtx.NodeImage
	cc.cluster.Spec.InitImage = cc.aisCtx.InitImage
	Expect(cc.k8sClient.Patch(cc.ctx, cc.cluster, patch)).Should(Succeed())
	By("Update cluster spec and wait for it to be 'Ready'")
	cc.waitForReadyCluster()
}

func (cc *clientCluster) getBaseParams() aisapi.BaseParams {
	cc.fetchLatestCluster()
	proxyURL := cc.getProxyURL()
	return aistutils.BaseAPIParams(proxyURL)
}

func (cc *clientCluster) fetchLatestCluster() {
	ais, err := cc.k8sClient.GetAIStoreCR(cc.ctx, cc.cluster.NamespacedName())
	Expect(err).To(BeNil())
	cc.cluster = ais
}

// Initialize AIS tutils to use the deployed cluster
func (cc *clientCluster) initClientAccess() {
	// Wait for all proxies
	proxyURLs := cc.getAllProxyURLs()
	for i := range proxyURLs {
		proxyURL := *proxyURLs[i]
		retries := 2
		for retries > 0 {
			err := aistutils.WaitNodeReady(proxyURL, &aistutils.WaitRetryOpts{
				MaxRetries: 12,
				Interval:   10 * time.Second,
			})
			if err == nil {
				break
			}
			retries--
			time.Sleep(5 * time.Second)
		}
		// Wait until the cluster has actually started (targets have registered).
		Expect(aistutils.InitCluster(proxyURL, aistutils.ClusterTypeK8s)).NotTo(HaveOccurred())
	}
}

func (cc *clientCluster) getProxyURL() (proxyURL string) {
	var ip string
	if cc.cluster.Spec.EnableExternalLB {
		ip = tutils.GetLoadBalancerIP(cc.ctx, cc.k8sClient, proxy.LoadBalancerSVCNSName(cc.cluster))
	} else {
		ip = tutils.GetRandomProxyIP(cc.ctx, cc.k8sClient, cc.cluster)
	}
	Expect(ip).NotTo(Equal(""))
	return fmt.Sprintf(urlTemplate, ip, cc.cluster.Spec.ProxySpec.ServicePort.String())
}

func (cc *clientCluster) getAllProxyURLs() (proxyURLs []*string) {
	var proxyIPs []string
	if cc.cluster.Spec.EnableExternalLB {
		proxyIPs = []string{tutils.GetLoadBalancerIP(cc.ctx, cc.k8sClient, proxy.LoadBalancerSVCNSName(cc.cluster))}
	} else {
		proxyIPs = tutils.GetAllProxyIPs(cc.ctx, cc.k8sClient, cc.cluster)
	}
	for _, ip := range proxyIPs {
		proxyURL := fmt.Sprintf(urlTemplate, ip, cc.cluster.Spec.ProxySpec.ServicePort.String())
		proxyURLs = append(proxyURLs, &proxyURL)
	}
	return proxyURLs
}

func (cc *clientCluster) destroyCleanupWithCallback(postDestroy func()) {
	cc.destroyAndCleanup()
	if postDestroy != nil {
		By("Running post-destroy callback")
		postDestroy()
	}
}

func (cc *clientCluster) destroyAndCleanup() {
	By(fmt.Sprintf("Destroying cluster %q", cc.cluster.Name))
	cc.destroyClusterOnly()
	if cc.pvs != nil {
		tutils.DestroyPV(context.Background(), cc.k8sClient, cc.pvs)
	}
}

func (cc *clientCluster) destroyClusterOnly() {
	tutils.DestroyCluster(context.Background(), cc.k8sClient, cc.cluster, clusterDestroyTimeout, clusterDestroyInterval)
}

func (cc *clientCluster) scale(targetOnly bool, factor int32) {
	By(fmt.Sprintf("Scaling cluster %q by %d", cc.cluster.Name, factor))
	cr, err := cc.k8sClient.GetAIStoreCR(cc.ctx, cc.cluster.NamespacedName())
	Expect(err).ShouldNot(HaveOccurred())
	patch := clientpkg.MergeFrom(cr.DeepCopy())
	if targetOnly {
		cr.Spec.TargetSpec.Size = aisapc.Ptr(cr.GetTargetSize() + factor)
	} else {
		cr.Spec.Size = aisapc.Ptr(*cr.Spec.Size + factor)
	}
	// Get current ready condition generation
	readyCond := tutils.GetClusterReadyCondition(cc.cluster)
	var readyGen int64
	if readyCond == nil {
		readyGen = 0
	} else {
		readyGen = readyCond.ObservedGeneration
	}
	Expect(cc.k8sClient.Patch(cc.ctx, cr, patch)).Should(Succeed())
	// Wait for the condition's generation to receive some update so we know reconciliation began
	// Otherwise, the cluster may be immediately ready
	tutils.WaitForReadyConditionChange(cc.ctx, cc.k8sClient, cr, readyGen, clusterUpdateTimeout, clusterUpdateInterval)
	cc.waitForReadyCluster()
	cc.initClientAccess()
}

func (cc *clientCluster) restart() {
	// Shutdown, ensure statefulsets exist and are size 0
	cc.setShutdownStatus(true)
	tutils.EventuallyPodsIsSize(cc.ctx, cc.k8sClient, cc.cluster, proxy.BasicLabels(cc.cluster), 0, clusterDestroyTimeout)
	tutils.EventuallyPodsIsSize(cc.ctx, cc.k8sClient, cc.cluster, target.BasicLabels(cc.cluster), 0, clusterDestroyTimeout)
	// Resume shutdown cluster, should become fully ready
	cc.setShutdownStatus(false)
	cc.waitForReadyCluster()
	cc.initClientAccess()
}

func (cc *clientCluster) setShutdownStatus(shutdown bool) {
	cr, err := cc.k8sClient.GetAIStoreCR(cc.ctx, cc.cluster.NamespacedName())
	Expect(err).ShouldNot(HaveOccurred())
	patch := clientpkg.MergeFrom(cr.DeepCopy())
	cr.Spec.ShutdownCluster = aisapc.Ptr(shutdown)
	err = cc.k8sClient.Patch(cc.ctx, cr, patch)
	Expect(err).ShouldNot(HaveOccurred())
}

func (cc *clientCluster) waitForResources() {
	tutils.CheckResExistence(cc.ctx, cc.cluster, cc.aisCtx, cc.k8sClient, true /*exists*/)
}

func (cc *clientCluster) waitForResourceDeletion() {
	tutils.CheckResExistence(cc.ctx, cc.cluster, cc.aisCtx, cc.k8sClient, false /*exists*/)
	tutils.CheckPVCDoesNotExist(cc.ctx, cc.cluster, cc.aisCtx, cc.k8sClient)
}

func (cc *clientCluster) printLogs() (err error) {
	cs, err := tutils.NewClientset()
	if err != nil {
		return fmt.Errorf("error creating clientset: %v", err)
	}

	clusterName := cc.cluster.Name
	clusterSelector := map[string]string{"app.kubernetes.io/name": clusterName}
	podList, err := cc.k8sClient.ListPods(cc.ctx, cc.cluster, clusterSelector)
	if err != nil {
		return fmt.Errorf("error listing pods for cluster %s: %v", clusterName, err)
	}
	for i := range podList.Items {
		pod := &podList.Items[i]
		opts := &corev1.PodLogOptions{Container: "ais-logs"}
		req := cs.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, opts)
		stream, err := req.Stream(cc.ctx)
		if err != nil {
			return fmt.Errorf("error opening log stream: %v", err)
		}
		defer stream.Close()
		fmt.Printf("Logs for pod %s in cluster %s:\n", pod.Name, clusterName)
		if _, err := io.Copy(os.Stdout, stream); err != nil {
			return fmt.Errorf("error printing logs for pod %s in cluster %s: %v", pod.Name, clusterName, err)
		}
	}
	return nil
}
