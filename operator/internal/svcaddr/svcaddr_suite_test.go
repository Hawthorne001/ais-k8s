/*
 * Copyright (c) 2026, NVIDIA CORPORATION. All rights reserved.
 */

package svcaddr

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestServiceAddresses(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service address suite")
}
