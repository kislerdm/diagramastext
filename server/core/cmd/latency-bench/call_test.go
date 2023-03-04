//go:build !unittest
// +build !unittest

package main

import (
	"context"
	"testing"

	"github.com/kislerdm/diagramastext/server/core/configuration"
	"github.com/kislerdm/diagramastext/server/core/openai"
)

func Benchmark_call(b *testing.B) {
	cfg, err := configuration.LoadDefaultConfig(context.Background())
	if err != nil {
		b.Fatal(err)
	}

	c, err := openai.NewClient(cfg.ModelInferenceConfig)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		call(context.TODO(), c)
	}
}
