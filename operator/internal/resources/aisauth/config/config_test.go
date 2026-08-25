/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

package config_test

import (
	"testing"
	"time"

	aisapc "github.com/NVIDIA/aistore/api/apc"
	aisauthn "github.com/NVIDIA/aistore/api/authn"
	"github.com/NVIDIA/aistore/cmn/cos"
	authv1alpha1 "github.com/ais-operator/api/aisauth/v1alpha1"
	authnconfig "github.com/ais-operator/internal/resources/aisauth/config"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenerateConfig(t *testing.T) {
	g := NewWithT(t)

	cfg := generateConfig(newTestAIStoreAuth())
	g.Expect(cfg.Net.HTTP.Port).To(HaveValue(Equal(52001)))
	g.Expect(cfg.Server.DB.Filepath).To(HaveValue(Equal("/etc/ais/authn/authn.db")))
}

func TestGenerateConfigUsesProvidedPaths(t *testing.T) {
	g := NewWithT(t)
	authn := newTestAIStoreAuth()
	authn.Spec.TLS = &authv1alpha1.TLSSpec{
		Certificate: &authv1alpha1.TLSCertificateConfig{
			IssuerRef: authv1alpha1.CertIssuerRef{Name: testIssuerName()},
		},
	}
	paths := authnconfig.Paths{
		Database:       "/custom/state/authn.db",
		TLSCertificate: "/custom/tls/server.crt",
		TLSKey:         "/custom/tls/server.key",
	}

	cfg := authnconfig.GenerateConfig(authn, paths)
	g.Expect(cfg.Server.DB.Filepath).To(HaveValue(Equal(paths.Database)))
	g.Expect(cfg.Net.HTTP.Certificate).To(HaveValue(Equal(paths.TLSCertificate)))
	g.Expect(cfg.Net.HTTP.Key).To(HaveValue(Equal(paths.TLSKey)))
}

func TestGenerateConfigSigningKey(t *testing.T) {
	g := NewWithT(t)
	bits := int32(4096)
	mode := aisauthn.SigningKeyModeExternal
	authn := newTestAIStoreAuth()
	authn.Spec.Config = &authv1alpha1.ConfigSpec{
		Auth: &authv1alpha1.ServerConfSpec{
			SigningKey: &authv1alpha1.SigningKeySpec{
				Bits: &bits,
				Mode: &mode,
			},
		},
	}

	cfg := generateConfig(authn)
	g.Expect(cfg.Server.SigningKey.Bits).To(HaveValue(Equal(4096)))
	g.Expect(cfg.Server.SigningKey.Mode).To(HaveValue(Equal(aisauthn.SigningKeyModeExternal)))
}

// A section the resource declares but leaves empty carries no settings, so it is not rendered.
func TestGenerateConfigOmitsEmptySections(t *testing.T) {
	t.Run("signing key", func(t *testing.T) {
		g := NewWithT(t)
		authn := newTestAIStoreAuth()
		authn.Spec.Config = &authv1alpha1.ConfigSpec{
			Auth: &authv1alpha1.ServerConfSpec{SigningKey: &authv1alpha1.SigningKeySpec{}},
		}

		g.Expect(generateConfig(authn).Server.SigningKey).To(BeNil())
	})

	t.Run("log", func(t *testing.T) {
		g := NewWithT(t)
		authn := newTestAIStoreAuth()
		authn.Spec.Config = &authv1alpha1.ConfigSpec{Log: &authv1alpha1.LogSpec{}}

		g.Expect(generateConfig(authn).Log).To(BeNil())
	})

	t.Run("timeout", func(t *testing.T) {
		g := NewWithT(t)
		authn := newTestAIStoreAuth()
		authn.Spec.Config = &authv1alpha1.ConfigSpec{Timeout: &authv1alpha1.TimeoutSpec{}}

		g.Expect(generateConfig(authn).Timeout).To(BeNil())
	})
}

func TestGenerateConfigIndividualAuthFields(t *testing.T) {
	t.Run("renders log config", func(t *testing.T) {
		g := NewWithT(t)
		level := int32(5)
		flushInterval := metav1.Duration{Duration: 11 * time.Second}
		authn := newTestAIStoreAuth()
		authn.Spec.Config = &authv1alpha1.ConfigSpec{
			Log: &authv1alpha1.LogSpec{
				Level:         &level,
				FlushInterval: &flushInterval,
			},
		}

		cfg := generateConfig(authn)
		g.Expect(cfg.Log.Level).To(HaveValue(Equal("5")))
		g.Expect(cfg.Log.FlushInterval).To(HaveValue(Equal(cos.Duration(11 * time.Second))))
	})

	t.Run("renders net config", func(t *testing.T) {
		g := NewWithT(t)
		port := int32(53001)
		externalURL := "https://authn.example.test:53001"
		authn := newTestAIStoreAuth()
		authn.Spec.Config = &authv1alpha1.ConfigSpec{
			Net: &authv1alpha1.NetSpec{
				ExternalURL: &externalURL,
				HTTP:        &authv1alpha1.HTTPConfSpec{Port: &port},
			},
		}

		cfg := generateConfig(authn)
		g.Expect(cfg.Net.ExternalURL).To(HaveValue(Equal(externalURL)))
		g.Expect(cfg.Net.HTTP.Port).To(HaveValue(Equal(53001)))
	})

	t.Run("renders tls paths", func(t *testing.T) {
		g := NewWithT(t)
		authn := newTestAIStoreAuth()
		authn.Spec.TLS = &authv1alpha1.TLSSpec{
			Certificate: &authv1alpha1.TLSCertificateConfig{
				IssuerRef: authv1alpha1.CertIssuerRef{Name: testIssuerName()},
			},
		}

		cfg := generateConfig(authn)
		g.Expect(cfg.Net.HTTP.UseHTTPS).To(HaveValue(BeTrue()))
		g.Expect(cfg.Net.HTTP.Certificate).To(HaveValue(Equal("/var/certs/tls.crt")))
		g.Expect(cfg.Net.HTTP.Key).To(HaveValue(Equal("/var/certs/tls.key")))
	})

	t.Run("renders expiration time", func(t *testing.T) {
		g := NewWithT(t)
		authn := newTestAIStoreAuth()
		authn.Spec.Config = &authv1alpha1.ConfigSpec{
			Auth: &authv1alpha1.ServerConfSpec{
				ExpirationTime: &metav1.Duration{Duration: 12 * time.Hour},
			},
		}

		cfg := generateConfig(authn)
		g.Expect(cfg.Server.Expire).To(HaveValue(Equal(cos.Duration(12 * time.Hour))))
	})

	t.Run("renders max token age", func(t *testing.T) {
		g := NewWithT(t)
		authn := newTestAIStoreAuth()
		authn.Spec.Config = &authv1alpha1.ConfigSpec{
			Auth: &authv1alpha1.ServerConfSpec{
				MaxTokenAge: &metav1.Duration{Duration: 72 * time.Hour},
			},
		}

		cfg := generateConfig(authn)
		g.Expect(cfg.Server.MaxTokenAge).To(HaveValue(Equal(cos.Duration(72 * time.Hour))))
	})

	t.Run("renders timeout", func(t *testing.T) {
		g := NewWithT(t)
		authn := newTestAIStoreAuth()
		authn.Spec.Config = &authv1alpha1.ConfigSpec{
			Timeout: &authv1alpha1.TimeoutSpec{
				DefaultTimeout: &metav1.Duration{Duration: 45 * time.Second},
			},
		}

		cfg := generateConfig(authn)
		g.Expect(cfg.Timeout.Default).To(HaveValue(Equal(cos.Duration(45 * time.Second))))
	})

	t.Run("includes db type", func(t *testing.T) {
		g := NewWithT(t)
		dbType := "BuntDB"
		authn := newTestAIStoreAuth()
		authn.Spec.Config = &authv1alpha1.ConfigSpec{
			Auth: &authv1alpha1.ServerConfSpec{
				DB: &authv1alpha1.DBSpec{Type: &dbType},
			},
		}

		cfg := generateConfig(authn)
		g.Expect(cfg.Server.DB.Type).To(HaveValue(Equal("BuntDB")))
		g.Expect(cfg.Server.DB.Filepath).To(HaveValue(Equal("/etc/ais/authn/authn.db")))
	})
}

// The rendered config carries operator-managed paths and whatever the resource sets, and nothing
// else. AuthN supplies the remaining values from its own defaults, which keeps the rendered config
// readable by server versions that do not share the operator's view of the config schema.
func TestGenerateConfigOmitsUnsetFields(t *testing.T) {
	t.Run("renders only operator-managed paths for a bare resource", func(t *testing.T) {
		g := NewWithT(t)

		g.Expect(marshalConfig(g, newTestAIStoreAuth())).To(Equal(map[string]any{
			"auth": map[string]any{
				"db": map[string]any{"filepath": "/etc/ais/authn/authn.db"},
			},
			"net": map[string]any{
				"http": map[string]any{"port": float64(52001)},
			},
		}))
	})

	t.Run("adds only the sections the resource configures", func(t *testing.T) {
		g := NewWithT(t)
		level := int32(4)
		authn := newTestAIStoreAuth()
		authn.Spec.Config = &authv1alpha1.ConfigSpec{
			Log: &authv1alpha1.LogSpec{Level: &level},
		}

		rendered := marshalConfig(g, authn)
		g.Expect(rendered).To(HaveKeyWithValue("log", map[string]any{"level": "4"}))
		g.Expect(rendered).NotTo(HaveKey("timeout"))
	})
}

func TestValidate(t *testing.T) {
	t.Run("accepts a bare resource", func(t *testing.T) {
		g := NewWithT(t)

		g.Expect(generateConfig(newTestAIStoreAuth()).Validate()).To(Succeed())
	})

	// Decoding into the AuthN config rejects unknown fields, so this fails if the operator
	// renders a setting AuthN does not recognize.
	t.Run("accepts every setting the operator renders", func(t *testing.T) {
		g := NewWithT(t)

		g.Expect(generateConfig(newFullyConfiguredAIStoreAuth()).Validate()).To(Succeed())
	})

	t.Run("rejects a token lifetime beyond the maximum token age", func(t *testing.T) {
		g := NewWithT(t)
		authn := newTestAIStoreAuth()
		authn.Spec.Config = &authv1alpha1.ConfigSpec{
			Auth: &authv1alpha1.ServerConfSpec{
				ExpirationTime: &metav1.Duration{Duration: 48 * time.Hour},
				MaxTokenAge:    &metav1.Duration{Duration: 24 * time.Hour},
			},
		}

		g.Expect(generateConfig(authn).Validate()).To(MatchError(ContainSubstring("max_token_age")))
	})

	t.Run("rejects a flush interval below the supported minimum", func(t *testing.T) {
		g := NewWithT(t)
		authn := newTestAIStoreAuth()
		authn.Spec.Config = &authv1alpha1.ConfigSpec{
			Log: &authv1alpha1.LogSpec{FlushInterval: &metav1.Duration{Duration: time.Second}},
		}

		g.Expect(generateConfig(authn).Validate()).To(MatchError(ContainSubstring("flush_interval")))
	})
}

func marshalConfig(g *WithT, authn *authv1alpha1.AIStoreAuth) map[string]any {
	confJSON, err := cos.JSON.MarshalToString(generateConfig(authn))
	g.Expect(err).NotTo(HaveOccurred())

	var rendered map[string]any
	g.Expect(cos.JSON.UnmarshalFromString(confJSON, &rendered)).To(Succeed())
	return rendered
}

func generateConfig(authn *authv1alpha1.AIStoreAuth) *authnconfig.Config {
	return authnconfig.GenerateConfig(authn, authnconfig.Paths{
		Database:       "/etc/ais/authn/authn.db",
		TLSCertificate: "/var/certs/tls.crt",
		TLSKey:         "/var/certs/tls.key",
	})
}

func newTestAIStoreAuth() *authv1alpha1.AIStoreAuth {
	return &authv1alpha1.AIStoreAuth{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ais-authn",
			Namespace: "ais",
		},
		Spec: authv1alpha1.AIStoreAuthSpec{
			Deployment: authv1alpha1.DeploymentSpec{
				Container: authv1alpha1.ContainerSpec{
					Image: "docker.io/aistorage/authn:v4.5",
				},
			},
		},
	}
}

// newFullyConfiguredAIStoreAuth sets every field the operator renders into AuthN config.
func newFullyConfiguredAIStoreAuth() *authv1alpha1.AIStoreAuth {
	authn := newTestAIStoreAuth()
	authn.Spec.TLS = &authv1alpha1.TLSSpec{
		Certificate: &authv1alpha1.TLSCertificateConfig{
			IssuerRef: authv1alpha1.CertIssuerRef{Name: testIssuerName()},
		},
	}
	authn.Spec.Config = &authv1alpha1.ConfigSpec{
		Auth: &authv1alpha1.ServerConfSpec{
			ExpirationTime: &metav1.Duration{Duration: 24 * time.Hour},
			MaxTokenAge:    &metav1.Duration{Duration: 72 * time.Hour},
			SigningKey: &authv1alpha1.SigningKeySpec{
				Bits: aisapc.Ptr(int32(4096)),
				Mode: aisapc.Ptr(aisauthn.SigningKeyModeExternal),
			},
			DB: &authv1alpha1.DBSpec{Type: aisapc.Ptr("BuntDB")},
		},
		Log: &authv1alpha1.LogSpec{
			Level:         aisapc.Ptr(int32(3)),
			FlushInterval: &metav1.Duration{Duration: 30 * time.Second},
		},
		Net: &authv1alpha1.NetSpec{
			ExternalURL: aisapc.Ptr("https://ais-authn.ais.svc.cluster.local:52001"),
			HTTP:        &authv1alpha1.HTTPConfSpec{Port: aisapc.Ptr(int32(52001))},
		},
		Timeout: &authv1alpha1.TimeoutSpec{
			DefaultTimeout: &metav1.Duration{Duration: 30 * time.Second},
		},
	}
	return authn
}

func testIssuerName() string {
	return "test-issuer"
}
