/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

// Package config renders AuthN server configuration from AIStoreAuth resources.
package config

import (
	"fmt"
	"strconv"

	aisapc "github.com/NVIDIA/aistore/api/apc"
	aisauthn "github.com/NVIDIA/aistore/api/authn"
	"github.com/NVIDIA/aistore/cmn/cos"
	authv1alpha1 "github.com/ais-operator/api/aisauth/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Paths contains the operator-managed file locations referenced by AuthN config.
type Paths struct {
	Database       string
	TLSCertificate string
	TLSKey         string
}

// GenerateConfig maps AIStoreAuth spec.config and spec.tls into the AuthN server configuration.
func GenerateConfig(authn *authv1alpha1.AIStoreAuth, paths Paths) *Config {
	conf := baseConfig(authn, paths)
	conf.applySpec(authn.Spec.Config)
	return conf
}

// Validate reports whether AuthN accepts the configuration, with absent fields checked as the
// defaults AuthN resolves on load.
func (c *Config) Validate() error {
	raw, err := cos.JSON.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal AuthN config: %w", err)
	}
	var full aisauthn.Config
	if err := cos.JSON.Unmarshal(raw, &full); err != nil {
		return fmt.Errorf("decode AuthN config: %w", err)
	}
	return full.Validate()
}

// baseConfig holds the operator-managed settings, which apply regardless of spec.config.
func baseConfig(authn *authv1alpha1.AIStoreAuth, paths Paths) *Config {
	conf := &Config{
		Net:    &NetConf{HTTP: &HTTPConf{Port: aisapc.Ptr(int(authn.ListenPort()))}},
		Server: &ServerConf{DB: &DatabaseConf{Filepath: aisapc.Ptr(paths.Database)}},
	}
	if authn.HasTLSEnabled() {
		conf.Net.HTTP.UseHTTPS = aisapc.Ptr(true)
		conf.Net.HTTP.Certificate = aisapc.Ptr(paths.TLSCertificate)
		conf.Net.HTTP.Key = aisapc.Ptr(paths.TLSKey)
	}
	return conf
}

// applySpec layers the settings from spec.config over the operator-managed ones.
func (c *Config) applySpec(specConf *authv1alpha1.ConfigSpec) {
	if specConf == nil {
		return
	}
	c.applyLogSpec(specConf.Log)
	c.applyNetSpec(specConf.Net)
	c.applyAuthSpec(specConf.Auth)
	c.applyTimeoutSpec(specConf.Timeout)
}

func (c *Config) applyLogSpec(logSpec *authv1alpha1.LogSpec) {
	if logSpec == nil || (logSpec.Level == nil && logSpec.FlushInterval == nil) {
		return
	}
	c.Log = &LogConf{FlushInterval: durationPtr(logSpec.FlushInterval)}
	if logSpec.Level != nil {
		c.Log.Level = aisapc.Ptr(strconv.FormatInt(int64(*logSpec.Level), 10))
	}
}

func (c *Config) applyNetSpec(netSpec *authv1alpha1.NetSpec) {
	if netSpec == nil {
		return
	}
	c.Net.ExternalURL = copyPtr(netSpec.ExternalURL)
}

func (c *Config) applyAuthSpec(authSpec *authv1alpha1.ServerConfSpec) {
	if authSpec == nil {
		return
	}
	c.Server.Expire = durationPtr(authSpec.ExpirationTime)
	c.Server.MaxTokenAge = durationPtr(authSpec.MaxTokenAge)
	if authSpec.DB != nil {
		c.Server.DB.Type = copyPtr(authSpec.DB.Type)
	}
	if authSpec.SigningKey != nil {
		c.Server.SigningKey = signingKeyConf(authSpec.SigningKey)
	}
}

func (c *Config) applyTimeoutSpec(timeoutSpec *authv1alpha1.TimeoutSpec) {
	if timeoutSpec == nil || timeoutSpec.DefaultTimeout == nil {
		return
	}
	c.Timeout = &TimeoutConf{Default: durationPtr(timeoutSpec.DefaultTimeout)}
}

func signingKeyConf(signingKeySpec *authv1alpha1.SigningKeySpec) *SigningKeyConf {
	if signingKeySpec.Bits == nil && signingKeySpec.Mode == nil {
		return nil
	}
	signingKey := &SigningKeyConf{Mode: copyPtr(signingKeySpec.Mode)}
	if signingKeySpec.Bits != nil {
		signingKey.Bits = aisapc.Ptr(int(*signingKeySpec.Bits))
	}
	return signingKey
}

// durationPtr converts an optional spec duration into the representation AuthN config uses.
func durationPtr(d *metav1.Duration) *cos.Duration {
	if d == nil {
		return nil
	}
	return aisapc.Ptr(cos.Duration(d.Duration))
}

// copyPtr returns a pointer to a copy of v, so the rendered config does not alias the resource.
func copyPtr[T any](v *T) *T {
	if v == nil {
		return nil
	}
	return aisapc.Ptr(*v)
}
