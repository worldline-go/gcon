package gcon_test

import (
	"context"
	"testing"
	"time"

	"github.com/worldline-go/gcon"
)

func TestConfig(t *testing.T) {
	t.Skip("skipping test")

	cfg := gcon.Config{
		Target: "localhost:50051",
		TLS: gcon.TLSConfig{
			Enabled: true,
		},
		Block:       true,
		DialTimeout: 5 * time.Second,
	}

	conn, err := cfg.Dial(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	_ = conn.Close()
}
