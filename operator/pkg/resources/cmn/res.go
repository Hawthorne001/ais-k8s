// Package cmn provides utilities for common AIS cluster resources
/*
 * Copyright (c) 2021-2024, NVIDIA CORPORATION. All rights reserved.
 */
package cmn

import (
	"path"

	aisapc "github.com/NVIDIA/aistore/api/apc"
	aisv1 "github.com/ais-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewAISVolumes(ais *aisv1.AIStore, daeType string) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: "config-mount",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "config-template-mount",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ais.Name + "-" + daeType,
					},
				},
			},
		},
		{
			Name: "config-global",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: globalConfigMapName(ais),
					},
				},
			},
		},
		{
			Name: "env-mount",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: path.Join(ais.Spec.HostpathPrefix, ais.Namespace, ais.Name, daeType+"_env"),
					Type: hostPathTypePtr(corev1.HostPathDirectoryOrCreate),
				},
			},
		},
		{
			Name: "state-mount",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: path.Join(ais.Spec.HostpathPrefix, ais.Namespace, ais.Name, daeType),
					Type: hostPathTypePtr(corev1.HostPathDirectoryOrCreate),
				},
			},
		},
		{
			Name: "statsd-config",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ais.Name + "-statsd",
					},
				},
			},
		},
	}
	if ais.Spec.AWSSecretName != nil {
		volumes = append(volumes, corev1.Volume{
			Name: "aws-creds",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: *ais.Spec.AWSSecretName,
				},
			},
		})
	}
	if ais.Spec.GCPSecretName != nil {
		volumes = append(volumes, corev1.Volume{
			Name: "gcp-creds",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: *ais.Spec.GCPSecretName,
				},
			},
		})
	}

	if ais.Spec.TLSSecretName != nil {
		volumes = append(volumes, corev1.Volume{
			Name: "tls-certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: *ais.Spec.TLSSecretName,
				},
			},
		})
	}

	if ais.Spec.LogsDirectory != "" {
		volumes = append(volumes, corev1.Volume{
			Name: "logs-dir",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: path.Join(ais.Spec.LogsDirectory, ais.Namespace, ais.Name, daeType),
					Type: hostPathTypePtr(corev1.HostPathDirectoryOrCreate),
				},
			},
		})
	}

	return volumes
}

func NewAISLivenessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"/bin/bash", "/var/ais_config/ais_liveness.sh"},
			},
		},
		InitialDelaySeconds: 90,
		PeriodSeconds:       5,
		FailureThreshold:    3,
		TimeoutSeconds:      5,
	}
}

func NewAISNodeLifecycle() *corev1.Lifecycle {
	return &corev1.Lifecycle{
		PreStop: &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"/bin/bash", "-c", "/usr/bin/pkill -SIGINT aisnode"},
			},
		},
	}
}

func NewAISVolumeMounts(spec *aisv1.AIStoreSpec, daeType string) []corev1.VolumeMount {
	hostMountSubPath := getHostMountSubPath(daeType)
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "config-mount",
			MountPath: "/var/ais_config",
		},
		{
			Name:      "config-global",
			MountPath: "/var/ais_config/ais.json",
			SubPath:   "ais.json",
		},
		{
			Name:      "config-global",
			MountPath: "/var/ais_config/ais_liveness.sh",
			SubPath:   "ais_liveness.sh",
		},
		{
			Name:      "config-global",
			MountPath: "/var/ais_config/ais_readiness.sh",
			SubPath:   "ais_readiness.sh",
		},
		{
			Name:        "env-mount",
			MountPath:   "/var/ais_env",
			SubPathExpr: hostMountSubPath,
		},
		{
			Name:        "state-mount",
			MountPath:   "/etc/ais",
			SubPathExpr: hostMountSubPath,
		},
		{
			Name:      "statsd-config",
			MountPath: "/var/statsd_config",
		},
	}

	if spec.AWSSecretName != nil {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "aws-creds",
			ReadOnly:  true,
			MountPath: "/root/.aws",
		})
	}
	if spec.GCPSecretName != nil {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "gcp-creds",
			ReadOnly:  true,
			MountPath: "/var/gcp",
		})
	}
	if spec.TLSSecretName != nil {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      "tls-certs",
			ReadOnly:  true,
			MountPath: "/var/certs",
		})
	}
	if spec.LogsDirectory != "" {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:        "logs-dir",
			MountPath:   "/var/log/ais",
			SubPathExpr: hostMountSubPath,
		})
	}

	return volumeMounts
}

func NewInitVolumeMounts(daeType string) []corev1.VolumeMount {
	hostMountSubPath := getHostMountSubPath(daeType)

	return []corev1.VolumeMount{
		{
			Name:      "config-mount",
			MountPath: "/var/ais_config",
		},
		{
			Name:      "config-template-mount",
			MountPath: "/var/ais_config_template",
		},
		{
			Name:      "config-global",
			MountPath: "/var/global_config",
		},
		{
			Name:        "env-mount",
			MountPath:   "/var/ais_env",
			SubPathExpr: hostMountSubPath,
		},
	}
}

func NewDaemonPorts(spec aisv1.DaemonSpec) []corev1.ContainerPort {
	var hostPort int32
	if spec.HostPort != nil {
		hostPort = *spec.HostPort
	}
	return []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: int32(spec.ServicePort.IntValue()),
			Protocol:      corev1.ProtocolTCP,
			HostPort:      hostPort,
		},
	}
}

func CreateAISAffinity(affinity *corev1.Affinity, podLabels map[string]string) *corev1.Affinity {
	// If we have no affinity defined in spec, define an empty one
	if affinity == nil {
		affinity = &corev1.Affinity{}
	}

	// If we have an affinity but no specific PodAntiAffinity, set it
	if affinity.PodAntiAffinity == nil {
		affinity.PodAntiAffinity = createPodAntiAffinity(podLabels)
	}

	return affinity
}

func createPodAntiAffinity(podLabels map[string]string) *corev1.PodAntiAffinity {
	// Pods matching podLabels may not be scheduled on the same hostname
	labelAffinity := corev1.PodAffinityTerm{
		LabelSelector: &metav1.LabelSelector{
			MatchLabels: podLabels,
		},
		TopologyKey: corev1.LabelHostname,
	}

	return &corev1.PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
			labelAffinity,
		},
	}
}

func getHostMountSubPath(daeType string) string {
	// Always use the pod name as sub path for targets, since target pods are bound to specific nodes
	if daeType == aisapc.Target {
		return "$(MY_POD)"
	}
	return ""
}

func hostPathTypePtr(v corev1.HostPathType) *corev1.HostPathType {
	return &v
}
