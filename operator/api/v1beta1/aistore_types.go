// Package contains declaration of AIS Kubernetes Custom Resource Definitions
/*
 * Copyright (c) 2021-2024, NVIDIA CORPORATION. All rights reserved.
 */
package v1beta1

import (
	aisapc "github.com/NVIDIA/aistore/api/apc"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type (
	ClusterCondition string
	ErrorReason      string
)

const (
	ConditionInitialized           ClusterCondition = "Initialized"
	ConditionInitializingLBService ClusterCondition = "InitializingLoadBalancerService"
	ConditionPendingLBService      ClusterCondition = "PendingLoadBalancerService"
	ConditionFailed                ClusterCondition = "Failed"
	ConditionCreated               ClusterCondition = "Created"
	ConditionReady                 ClusterCondition = "Ready"
	ConditionUpgrading             ClusterCondition = "Upgrading"
	// TODO: Add more states, eg. Terminating etc.

	// Condition types
	ReconcilerError         string = "ReconcilerError"
	ReconcilerSuccess       string = "ReconcilerSuccess"
	ReconcilerSuccessReason string = "LastReconcileCycleSucceded"

	// ErrorReason
	ReasonUnknown         ErrorReason = "Unknown"
	IncompatibleSpecError ErrorReason = "IncompatibleSpecError"
	RBACManagementError   ErrorReason = "RBACError"
	ProxyCreationError    ErrorReason = "ProxyCreationError"
	TargetCreationError   ErrorReason = "TargetCreationError"
	InstanceDeletionError ErrorReason = "InstanceDeletionError"
	ConfigChangeError     ErrorReason = "ConfigChangeError"
	ConfigBuildError      ErrorReason = "ConfigBuildError"
	OwnerReferenceError   ErrorReason = "OwnerReferenceError"
	ExternalServiceError  ErrorReason = "ExternalService"
	ResourceCreationError ErrorReason = "ResourceCreationError"
	ResourceFetchError    ErrorReason = "ResouceFetchError" // failed to fetch a resource using K8s API
	ResourceUpdateError   ErrorReason = "ResourceUpdateError"

	defaultClusterDomain = "cluster.local"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// IMPORTANT: Run "make" to regenerate code after modifying this file

// AIStoreSpec defines the desired state of AIStore
type AIStoreSpec struct {
	// Size of the cluster i.e. number of proxies and number of targets.
	// This can be changed by specifying size in either `proxySpec` or `targetSpec`.
	// +optional
	Size           *int32          `json:"size"`
	NodeImage      string          `json:"nodeImage"` // docker image of aisnode
	InitImage      string          `json:"initImage"` // init image for nodes
	HostpathPrefix string          `json:"hostpathPrefix"`
	ConfigToUpdate *ConfigToUpdate `json:"configToUpdate,omitempty"`
	// Map of primary host to comma-separated string of all hosts for multi-home
	// +optional
	HostnameMap map[string]string `json:"hostnameMap,omitempty"`
	// Commma-separated list of names of additional network attachment definitions to attach to each pod
	// +optional
	NetAttachment *string `json:"networkAttachment,omitempty"`

	ProxySpec  DaemonSpec `json:"proxySpec"`  // spec for proxy
	TargetSpec TargetSpec `json:"targetSpec"` // spec for target

	// ShutdownCluster can be set true if the desired state of the cluster is shutdown with a future restart expected
	// When enabled, the operator will gracefully shut down the AIS cluster and scale cluster size to 0
	// No data or configuration will be deleted
	// +optional
	ShutdownCluster *bool `json:"shutdownCluster,omitempty"`

	// CleanupData determines whether to clean up PVCs and user data (including buckets and objects) when the cluster is decommissioned.
	// The reclamation of PVs linked to the PVCs depends on the PV reclaim policy or the default policy of the associated StorageClass.
	// This field is relevant only if you are deleting the CR (leading to decommissioning of the cluster).
	// +optional
	CleanupData *bool `json:"cleanupData,omitempty"`

	// Defines if AIS daemons should expose prometheus metrics
	// +optional
	EnablePromExporter *bool `json:"enablePromExporter,omitempty"`

	// Defines the cluster domain name for DNS. Default: cluster.local.
	// +optional
	ClusterDomain *string `json:"clusterDomain,omitempty"`

	// Secret name containing GCP credentials
	// +optional
	GCPSecretName *string `json:"gcpSecretName,omitempty"`

	// Secret name containing AWS credentials
	// +optional
	AWSSecretName *string `json:"awsSecretName,omitempty"`

	// Logs directory on host to store AIS logs
	// +optional
	LogsDirectory string `json:"logsDir,omitempty"`

	// Secret name containing TLS cert/key
	// +optional
	TLSSecretName *string `json:"tlsSecretName,omitempty"`

	// ImagePullScerets is an optional list of references to secrets in the same namespace to pull container images of AIS Daemons
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// Deprecated: use TargetSpec.DisablePodAntiAffinity
	// DisablePodAntiAffinity, if set allows more than one target/proxy daemon pods to be scheduled on same K8s node.
	// +optional
	DisablePodAntiAffinity *bool `json:"disablePodAntiAffinity,omitempty"`

	// EnableExternalLB, if set, enables external access to AIS cluster using LoadBalancer service
	EnableExternalLB bool `json:"enableExternalLB"`
}

// AIStoreStatus defines the observed state of AIStore
type AIStoreStatus struct {
	// Represents the observations of a AIStores's current state.
	// Known .status.conditions.type are: "Initialized", "Created", and "Ready"
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions"`
	// +optional
	State ClusterCondition `json:"state"`
	// +optional
	ConsecutiveErrorCount int `json:"consecutive_error_count"` // number of times an error occurred
}

// ServiceSpec defines the specs of AIS Gateways
type ServiceSpec struct {
	ServicePort      intstr.IntOrString `json:"servicePort"`
	PublicPort       intstr.IntOrString `json:"portPublic"` // port of PublicNet
	IntraControlPort intstr.IntOrString `json:"portIntraControl"`
	IntraDataPort    intstr.IntOrString `json:"portIntraData"`
}

// NodeSpec defines the specs for AIS Daemon pods/containers
type DaemonSpec struct {
	ServiceSpec `json:",inline"`
	// SecurityContext holds pod-level security attributes and common container settings for AIS Daemon (proxy/target) object.
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`

	// Size holds number of AIS Daemon (proxy/target) replicas.
	// Overrides value present in `AIStore` spec.
	// +optional
	Size *int32 `json:"size"`

	// ContainerSecurity holds the secrity context for AIS Daemon containers.
	// +optional
	ContainerSecurity *corev1.SecurityContext `json:"capabilities,omitempty"`
	// Affinity  - AIS Daemon pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`
	// NodeSelector -  which must match a node's labels for the AIS Daemon pod to be scheduled on that node.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations - list of tolerations for AIS Daemon pod
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// HostPort - host port to use for hostnetworking
	// +optional
	HostPort *int32 `json:"hostPort,omitempty"`
}

type TargetSpec struct {
	DaemonSpec `json:",inline"`
	Mounts     []Mount `json:"mounts"`
	// Deprecated: AllowSharedOrNoDisks - disables FsID and mountpath disks validation on target nodes
	// NOT recommended for production deployments
	// Use Mount.Label instead
	// +optional
	AllowSharedOrNoDisks *bool `json:"allowSharedNoDisks,omitempty"`

	// DisablePodAntiAffinity allows more than one target pod to be scheduled on same K8s node.
	// +optional
	DisablePodAntiAffinity *bool `json:"disablePodAntiAffinity,omitempty"`

	// hostNetwork - if set to true, the AIS Daemon pods for target are created in the host's network namespace (used for multihoming)
	// +optional
	HostNetwork *bool `json:"hostNetwork,omitempty"`
}

type Mount struct {
	Path         string                `json:"path"`
	Size         resource.Quantity     `json:"size"`
	StorageClass *string               `json:"storageClass,omitempty"` // storage class for volume resource
	Selector     *metav1.LabelSelector `json:"selector,omitempty"`     // selector for choosing PVs
	// Mountpath labels can be used for mapping mountpaths to disks, enabling disk sharing,
	// defining storage classes for bucket-specific storage, and allowing user-defined mountpath
	// grouping for capacity and storage class differentiation
	Label *string `json:"label,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AIStore is the Schema for the aistores API
type AIStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AIStoreSpec   `json:"spec,omitempty"`
	Status AIStoreStatus `json:"status,omitempty"`
}

// AddOrUpdateCondition is used to add a new/update an existing condition type.
func (ais *AIStore) AddOrUpdateCondition(c metav1.Condition) {
	c.LastTransitionTime = metav1.Now()
	c.ObservedGeneration = ais.GetGeneration()
	for i, condition := range ais.Status.Conditions {
		if c.Type == condition.Type {
			ais.Status.Conditions[i] = c
			return
		}
	}
	ais.Status.Conditions = append(ais.Status.Conditions, c)
}

// GetLastCondition returns the last condition based on the condition timestamp.
// Return false if no condition is present.
func (ais *AIStore) GetLastCondition() (latest metav1.Condition, exists bool) {
	if len(ais.Status.Conditions) == 0 {
		return
	}
	exists = true
	latest = ais.Status.Conditions[0]
	lastTime := latest.LastTransitionTime
	for i, condition := range ais.Status.Conditions {
		if i == 0 {
			continue
		}
		if lastTime.Before(&condition.LastTransitionTime) {
			latest = condition
			lastTime = condition.LastTransitionTime
		}
	}
	return
}

// SetConditionInitialized add a new condition type `Initialized` and sets it to `True`
func (ais *AIStore) SetConditionInitialized() {
	ais.AddOrUpdateCondition(metav1.Condition{
		Type:    ConditionInitialized.Str(),
		Status:  metav1.ConditionTrue,
		Reason:  ConditionInitialized.Str(),
		Message: "Success initializing cluster",
	})
}

// SetConditionCreated add a new condition type `Created` and sets it to `True`
func (ais *AIStore) SetConditionCreated() {
	ais.AddOrUpdateCondition(metav1.Condition{
		Type:    ConditionCreated.Str(),
		Status:  metav1.ConditionTrue,
		Reason:  ConditionCreated.Str(),
		Message: "Success creating AIS cluster",
	})
}

// SetConditionReady add a new condition type `Ready` and sets it to `True`
func (ais *AIStore) SetConditionReady() {
	ais.AddOrUpdateCondition(metav1.Condition{
		Type:    ConditionReady.Str(),
		Status:  metav1.ConditionTrue,
		Reason:  ConditionReady.Str(),
		Message: "Cluster is ready",
	})
}

// UnsetConditionReady add/updates condition setting type `Ready` to `False`
// reason - tag why the condition is being set to `False`.
// message - a human readable message indicating details about state change.
func (ais *AIStore) UnsetConditionReady(reason, message string) {
	ais.AddOrUpdateCondition(metav1.Condition{
		Type:    ConditionReady.Str(),
		Status:  metav1.ConditionFalse,
		Reason:  reason,
		Message: message,
	})
}

// SetConditionError sets records error occurred in reconciler loop
func (ais *AIStore) SetConditionError(reason ErrorReason, err error) {
	if err == nil {
		return
	}
	ais.AddOrUpdateCondition(metav1.Condition{
		Type:    ReconcilerError,
		Status:  metav1.ConditionTrue,
		Reason:  reason.Str(),
		Message: err.Error(),
	})
}

func (ais *AIStore) IncErrorCount()   { ais.Status.ConsecutiveErrorCount++ }
func (ais *AIStore) ResetErrorCount() { ais.Status.ConsecutiveErrorCount = 0 }
func (ais *AIStore) SetConditionSuccess() {
	ais.Status.ConsecutiveErrorCount = 0
	ais.AddOrUpdateCondition(metav1.Condition{
		Type:   ReconcilerSuccess,
		Status: metav1.ConditionTrue,
		Reason: ReconcilerSuccessReason,
	})
}

func (ais *AIStore) getCondition(conditionType string) (metav1.Condition, bool) {
	for _, condition := range ais.Status.Conditions {
		if condition.Type == conditionType {
			return condition, true
		}
	}
	return metav1.Condition{}, false
}

// IsConditionTrue checks if the `Status` for given type is set to true
func (ais *AIStore) IsConditionTrue(conditionType string) (isTrue bool) {
	condition, ok := ais.getCondition(conditionType)
	if !ok {
		return
	}
	isTrue = condition.Status == metav1.ConditionTrue
	return
}

func (ais *AIStore) SetState(state ClusterCondition) {
	ais.Status.State = state
}

func (ais *AIStore) HasState(state ClusterCondition) bool {
	return ais.Status.State == state
}

func (ais *AIStore) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      ais.Name,
		Namespace: ais.Namespace,
	}
}

func (ais *AIStore) GetClusterDomain() string {
	if ais.Spec.ClusterDomain == nil {
		return defaultClusterDomain
	}
	return *ais.Spec.ClusterDomain
}

func (ais *AIStore) ProxyStatefulSetName() string {
	return ais.Name + "-" + aisapc.Proxy
}

func (ais *AIStore) DefaultPrimaryName() string {
	return ais.ProxyStatefulSetName() + "-0"
}

func (ais *AIStore) GetProxySize() int32 {
	if ais.Spec.ProxySpec.Size != nil {
		return *ais.Spec.ProxySpec.Size
	}
	return *ais.Spec.Size
}

func (ais *AIStore) GetTargetSize() int32 {
	if ais.Spec.TargetSpec.Size != nil {
		return *ais.Spec.TargetSpec.Size
	}
	return *ais.Spec.Size
}

func (ais *AIStore) ShouldShutdown() bool {
	return ais.Spec.ShutdownCluster != nil && *ais.Spec.ShutdownCluster
}

func (ais *AIStore) AllowTargetSharedNodes() bool {
	allowSharedNodes := ais.Spec.TargetSpec.DisablePodAntiAffinity != nil && *ais.Spec.TargetSpec.DisablePodAntiAffinity
	//nolint:all
	deprecatedAllow := ais.Spec.DisablePodAntiAffinity != nil && *ais.Spec.DisablePodAntiAffinity
	// Backwards compatible check -- allow if either is true
	return allowSharedNodes || deprecatedAllow
}

// +kubebuilder:object:root=true

// AIStoreList contains a list of AIStore
type AIStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AIStore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AIStore{}, &AIStoreList{})
}

////////////////////////
//    ErrorReason     //
///////////////////////

func (e ErrorReason) Equals(value string) bool {
	return string(e) == value
}

func (e ErrorReason) Str() string {
	return string(e)
}

/////////////////////////////////
//     ClusterCondition       //
///////////////////////////////

func (c ClusterCondition) Str() string {
	return string(c)
}
