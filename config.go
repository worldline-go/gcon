package gcon

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/twmb/tlscfg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	// Target is the address of the gRPC server, as "localhost:50051"
	Target string `cfg:"target"`
	// Authority sets the grpc.WithAuthority.
	Authority string `cfg:"authority"`

	// TLS is the TLS configuration for the connection.
	TLS TLSConfig `cfg:"tls"`
	// Block will block the Dial method until connection is established.
	Block       bool          `cfg:"block"        default:"true"`
	DialTimeout time.Duration `cfg:"dial_timeout" default:"5s"`
}

type TLSConfig struct {
	Enabled            bool `cfg:"enabled"`
	InsecureSkipVerify bool `cfg:"insecure_skip_verify"`
	// CertFile is the path to the client's TLS certificate.
	// Should be use with KeyFile.
	CertFile string `cfg:"cert_file"`
	// KeyFile is the path to the client's TLS key.
	// Should be use with CertFile.
	KeyFile string `cfg:"key_file"`
	// CAFile is the path to the CA certificate.
	// If empty, the server's root CA set will be used.
	CAFile string `cfg:"ca_file"`
	// ServerName is the server name used to verify the hostname on the returned certificates from the server.
	ServerName string `cfg:"server_name"`
}

// Generate returns a tls.Config based on the TLSConfig.
func (t TLSConfig) Generate() (*tls.Config, error) {
	if !t.Enabled {
		return nil, fmt.Errorf("TLS is not enabled")
	}

	opts := []tlscfg.Opt{}

	// load client cert
	if t.CertFile != "" && t.KeyFile != "" {
		opts = append(opts, tlscfg.WithDiskKeyPair(t.CertFile, t.KeyFile))
	}

	// load CA cert
	opts = append(opts, tlscfg.WithSystemCertPool())
	if t.CAFile != "" {
		opts = append(opts, tlscfg.WithDiskCA(t.CAFile, tlscfg.ForClient))
	}

	cfg, err := tlscfg.New(opts...)
	if err != nil {
		return nil, err
	}

	if t.InsecureSkipVerify {
		cfg.InsecureSkipVerify = true
	}

	if t.ServerName != "" {
		cfg.ServerName = t.ServerName
	}

	return cfg, nil
}

// Dial dials the target gRPC server.
func (c Config) Dial(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	options, err := c.Options()
	if err != nil {
		return nil, fmt.Errorf("failed to get dial options: %w", err)
	}

	opts = append(options, opts...)

	if c.DialTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.DialTimeout)
		defer cancel()
	}

	return grpc.DialContext(ctx, c.Target, opts...)
}

// Options returns the gRPC dial options based on the Config.
func (c Config) Options() ([]grpc.DialOption, error) {
	opts := []grpc.DialOption{}

	if c.Authority != "" {
		opts = append(opts, grpc.WithAuthority(c.Authority))
	}

	if c.TLS.Enabled {
		tlsCfg, err := c.TLS.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate TLS config: %w", err)
		}

		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if c.Block {
		opts = append(opts, grpc.WithBlock())
	}

	return opts, nil
}
