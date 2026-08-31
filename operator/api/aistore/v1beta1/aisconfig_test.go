/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

package v1beta1

import (
	"testing"

	aisapc "github.com/NVIDIA/aistore/api/apc"
	. "github.com/onsi/gomega"
)

func TestConfigToUpdateRequiresClientAuth(t *testing.T) {
	tests := []struct {
		name string
		conf *ConfigToUpdate
		want bool
	}{
		{name: "nil config"},
		{name: "no auth section", conf: &ConfigToUpdate{}},
		{name: "auth section without enabled", conf: &ConfigToUpdate{Auth: &AuthConfToUpdate{}}},
		{
			name: "signing method without enabled",
			conf: &ConfigToUpdate{Auth: &AuthConfToUpdate{
				Signature: &AuthSignatureConfToUpdate{Method: aisapc.Ptr(SigningKeyMethodHMAC)},
			}},
		},
		{
			name: "auth explicitly disabled",
			conf: &ConfigToUpdate{Auth: &AuthConfToUpdate{Enabled: aisapc.Ptr(false)}},
		},
		{
			name: "auth explicitly enabled",
			conf: &ConfigToUpdate{Auth: &AuthConfToUpdate{Enabled: aisapc.Ptr(true)}},
			want: true,
		},
		{
			name: "client auth not required",
			conf: &ConfigToUpdate{Auth: &AuthConfToUpdate{ClientAuthRequired: aisapc.Ptr(false)}},
		},
		{
			name: "client auth required",
			conf: &ConfigToUpdate{Auth: &AuthConfToUpdate{ClientAuthRequired: aisapc.Ptr(true)}},
			want: true,
		},
		{
			name: "client_auth_required false overrides deprecated enabled",
			conf: &ConfigToUpdate{Auth: &AuthConfToUpdate{
				Enabled:            aisapc.Ptr(true),
				ClientAuthRequired: aisapc.Ptr(false),
			}},
		},
		{
			name: "client_auth_required true overrides deprecated enabled",
			conf: &ConfigToUpdate{Auth: &AuthConfToUpdate{
				Enabled:            aisapc.Ptr(false),
				ClientAuthRequired: aisapc.Ptr(true),
			}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(subT *testing.T) {
			g := NewWithT(subT)
			g.Expect(tt.conf.RequiresClientAuth()).To(Equal(tt.want))
		})
	}
}
