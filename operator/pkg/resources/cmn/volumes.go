// Package cmn provides utilities for common AIS cluster resources
/*
 * Copyright (c) 2021-2024, NVIDIA CORPORATION. All rights reserved.
 */
package cmn

import (
	"path"
	"strings"

	aisapc "github.com/NVIDIA/aistore/api/apc"
	"github.com/ais-operator/api/v1beta1"
	"github.com/ais-operator/pkg/resources/statsd"
	csiapis "github.com/cert-manager/csi-driver/pkg/apis"
	csiapisv1 "github.com/cert-manager/csi-driver/pkg/apis/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

const (
	// StateDir Container-internal location of configs and current state of the aisnode
	StateDir = "/etc/ais"
	// InitConfTemplateDir Container-internal location of config template, mounted from the config map
	InitConfTemplateDir = "/var/ais_config_template"
	// AisConfigDir Container-internal location of initial config, written by init container and used at aisnode start
	AisConfigDir      = "/var/ais_config"
	LogsDir           = "/var/log/ais"
	StatsDDir         = "/var/statsd_config"
	InitGlobalConfDir = "/var/global_config"

	// Container mount locations for cloud provider configs
	DefaultGCPDir = "/var/gcp"
	DefaultAWSDir = "/root/.aws"
	DefaultOCIDir = "/root/.oci"

	// Other container mount locations
	certsDir  = "/var/certs"
	tracesDir = "/var/traces"

	hostnameMapFileName = "hostname_map.json"
	AISGlobalConfigName = "ais.json"
	AISLocalConfigName  = "ais_local.json"
	AISDataPVC          = "target_pvcs.json"

	StatsDVolume         = "statsd-config"
	configTemplateVolume = "config-template"
	configVolume         = "config-mount"
	configGlobalVolume   = "config-global"
	stateVolume          = "state-mount"
	awsSecretVolume      = "aws-creds"
	gcpSecretVolume      = "gcp-creds" //nolint:gosec // This is not really credential.
	ociSecretVolume      = "oci-creds"
	tlsSecretVolume      = "tls-certs"
	tracingSecretVolume  = "tracing-token"
	logsVolume           = "logs-dir"
)

func NewAISVolumes(ais *v1beta1.AIStore, daeType string) []v1.Volume {
	volumes := []v1.Volume{
		{
			Name: configTemplateVolume,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: AISConfigMapName(ais, daeType),
					},
				},
			},
		},
		{
			Name: configVolume,
			VolumeSource: v1.VolumeSource{
				EmptyDir: &v1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: configGlobalVolume,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: globalConfigMapName(ais),
					},
				},
			},
		},
		{
			Name: StatsDVolume,
			VolumeSource: v1.VolumeSource{
				ConfigMap: &v1.ConfigMapVolumeSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: statsd.ConfigMapName(ais),
					},
				},
			},
		},
		newLogsVolume(ais, daeType),
	}

	// Only create hostpath volumes if no storage class is provided for state
	if ais.Spec.StateStorageClass == nil {
		hostpathVolumes := []v1.Volume{
			{
				Name: stateVolume,
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						//nolint:all
						Path: path.Join(*ais.Spec.HostpathPrefix, ais.Namespace, ais.Name, daeType),
						Type: aisapc.Ptr(v1.HostPathDirectoryOrCreate),
					},
				},
			},
		}
		volumes = append(volumes, hostpathVolumes...)
	}

	if daeType == aisapc.Target {
		volumes = append(volumes, newCloudVolumes(ais)...)
	}

	if ais.Spec.TLSCertManagerIssuerName != nil {
		name := ais.Name + "-" + daeType
		volumes = append(volumes, v1.Volume{
			Name: tlsSecretVolume,
			VolumeSource: v1.VolumeSource{
				CSI: &v1.CSIVolumeSource{
					Driver: csiapis.GroupName,
					VolumeAttributes: map[string]string{
						csiapisv1.IssuerNameKey: *ais.Spec.TLSCertManagerIssuerName,
						csiapisv1.CommonNameKey: name + ".${POD_NAMESPACE}",
						csiapisv1.DNSNamesKey: strings.Join(
							[]string{
								"${POD_NAME}.${POD_NAMESPACE}.svc." + ais.GetClusterDomain(),
								name + ".${POD_NAMESPACE}.svc." + ais.GetClusterDomain(),
								name + ".${POD_NAMESPACE}.svc",
								name,
							},
							","),
					},
					ReadOnly: aisapc.Ptr(true),
				},
			},
		})
	} else if ais.Spec.TLSSecretName != nil {
		volumes = append(volumes, v1.Volume{
			Name: tlsSecretVolume,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: *ais.Spec.TLSSecretName,
				},
			},
		})
	}

	if ais.Spec.TracingTokenSecretName != nil {
		volumes = append(volumes, v1.Volume{
			Name: tracingSecretVolume,
			VolumeSource: v1.VolumeSource{
				Secret: &v1.SecretVolumeSource{
					SecretName: *ais.Spec.TracingTokenSecretName,
				},
			},
		})
	}
	return volumes
}

func newCloudVolumes(ais *v1beta1.AIStore) []v1.Volume {
	var volumes []v1.Volume

	type cloudSecret struct {
		namePtr    *string
		volumeName string
	}

	secrets := []cloudSecret{
		{ais.Spec.AWSSecretName, awsSecretVolume},
		{ais.Spec.GCPSecretName, gcpSecretVolume},
		{ais.Spec.OCISecretName, ociSecretVolume},
	}

	for _, secret := range secrets {
		if secret.namePtr != nil {
			volumes = append(volumes, v1.Volume{
				Name: secret.volumeName,
				VolumeSource: v1.VolumeSource{
					Secret: &v1.SecretVolumeSource{
						SecretName: *secret.namePtr,
					},
				},
			})
		}
	}

	return volumes
}

func newLogsVolume(ais *v1beta1.AIStore, daeType string) v1.Volume {
	if ais.Spec.LogsDirectory != "" {
		return v1.Volume{
			Name: logsVolume,
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: path.Join(ais.Spec.LogsDirectory, ais.Namespace, ais.Name, daeType),
					Type: aisapc.Ptr(v1.HostPathDirectoryOrCreate),
				},
			},
		}
	}
	return v1.Volume{
		Name: logsVolume,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
}

func NewAISVolumeMounts(ais *v1beta1.AIStore, daeType string) []v1.VolumeMount {
	spec := &ais.Spec
	volumeMounts := []v1.VolumeMount{
		{
			Name:      configVolume,
			MountPath: AisConfigDir,
		},
		{
			Name:      StatsDVolume,
			MountPath: StatsDDir,
		},
		newLogsVolumeMount(daeType),
	}

	if spec.StateStorageClass != nil {
		volumeName := getStatePVCName(ais)
		dynamicMounts := []v1.VolumeMount{
			{
				Name:      volumeName,
				MountPath: StateDir,
			},
		}
		volumeMounts = append(volumeMounts, dynamicMounts...)
	} else {
		hostMountSubPath := getHostMountSubPath(daeType)
		hostMounts := []v1.VolumeMount{
			{
				Name:        stateVolume,
				MountPath:   StateDir,
				SubPathExpr: hostMountSubPath,
			},
		}
		volumeMounts = append(volumeMounts, hostMounts...)
	}

	if daeType == aisapc.Target {
		volumeMounts = appendCloudVolumeMounts(spec, volumeMounts)
	}

	if spec.TLSCertManagerIssuerName != nil || spec.TLSSecretName != nil {
		volumeMounts = appendSimpleReadOnlyMount(volumeMounts, tlsSecretVolume, certsDir)
	}
	if spec.TracingTokenSecretName != nil {
		volumeMounts = appendSimpleReadOnlyMount(volumeMounts, tracingSecretVolume, tracesDir)
	}
	return volumeMounts
}

func appendCloudVolumeMounts(spec *v1beta1.AIStoreSpec, mounts []v1.VolumeMount) []v1.VolumeMount {
	type cloudConfig struct {
		secretName *string
		defaultDir string
		volumeName string
	}

	configs := []cloudConfig{
		{spec.AWSSecretName, DefaultAWSDir, awsSecretVolume},
		{spec.GCPSecretName, DefaultGCPDir, gcpSecretVolume},
		{spec.OCISecretName, DefaultOCIDir, ociSecretVolume},
	}

	for _, cfg := range configs {
		if cfg.secretName != nil {
			mounts = appendSimpleReadOnlyMount(mounts, cfg.volumeName, cfg.defaultDir)
		}
	}
	return mounts
}

func appendSimpleReadOnlyMount(mounts []v1.VolumeMount, name, mountPath string) []v1.VolumeMount {
	return append(mounts, v1.VolumeMount{
		Name:      name,
		ReadOnly:  true,
		MountPath: mountPath,
	})
}

func newLogsVolumeMount(daeType string) v1.VolumeMount {
	return v1.VolumeMount{
		Name:        logsVolume,
		MountPath:   LogsDir,
		SubPathExpr: getHostMountSubPath(daeType),
	}
}

func NewInitVolumeMounts() []v1.VolumeMount {
	volumeMounts := []v1.VolumeMount{
		{
			Name:      configTemplateVolume,
			MountPath: InitConfTemplateDir,
		},
		{
			Name:      configVolume,
			MountPath: AisConfigDir,
		},
		{
			Name:      configGlobalVolume,
			MountPath: InitGlobalConfDir,
		},
	}
	return volumeMounts
}

func getHostMountSubPath(daeType string) string {
	// Always use the pod name as sub path for targets, since target pods are bound to specific nodes
	if daeType == aisapc.Target {
		return "$(MY_POD)"
	}
	return ""
}
