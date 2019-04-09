package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	g "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"sync/atomic"
	"time"
	cfg "xianhetian.com/tendermint/p2p/grpc/config"
)

type Client struct {
	g.ClientConn
	Config *cfg.ClientConfig
}

func NewClient() *Client {
	return &Client{
		Config: cfg.DefaultClientConfig(),
	}
}

func (c *Client) Conn(address ...string) ([]g.ClientConn, error) {
	var opts []g.DialOption
	var clientConns []g.ClientConn
	opts = append(opts, c.setupSecure()...)
	opts = append(opts, c.setupKeepalive()...)
	opts = append(opts, g.WithDefaultCallOptions(g.MaxCallRecvMsgSize(c.Config.MaxRecvMsgSize), g.MaxCallSendMsgSize(c.Config.MaxSendMsgSize)))
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(c.Config.Timeout))
	defer cancel()
	switch len(address) {
	case 0:
		conn, err := g.DialContext(ctx, c.Config.Addr, opts...)
		if err != nil {
			return nil, err
		}
		c.ClientConn = *conn
	case 1:
		conn, err := g.DialContext(ctx, address[0], opts...)
		if err != nil {
			return nil, err
		}
		c.ClientConn = *conn
	default:
		for _, a := range address {
			i, err := g.DialContext(ctx, a, opts...)
			if err != nil {
				return nil, err
			}
			clientConns = append(clientConns, *i)
		}
	}
	return clientConns, nil
}

func (c *Client) setupKeepalive() []g.DialOption {
	if !c.Config.UseKeepalive {
		return nil
	}
	var opts []g.DialOption
	kap := keepalive.ClientParameters{
		Time:                c.Config.Interval,
		Timeout:             c.Config.AliveTimeout,
		PermitWithoutStream: c.Config.PermitWithoutStream,
	}
	return append(opts, g.WithKeepaliveParams(kap), g.WithBlock())
}

func (c *Client) setupSecure() []g.DialOption {
	var opts []g.DialOption
	var SecureCertificate atomic.Value
	if !c.Config.UseTLS && c.Config.PreCreds == nil {
		opts = append(opts, g.WithInsecure())
		return opts
	}
	if c.Config.PreCreds != nil {
		opts = append(opts, g.WithPerRPCCredentials(c.Config.PreCreds))
	}
	if !c.Config.UseTLS {
		return opts
	}
	cert, err := tls.X509KeyPair(c.Config.Certificate, c.Config.Key)
	if err != nil {
		panic(err)
	}
	SecureCertificate.Store(cert)
	c.Config.GetCert = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		cert := SecureCertificate.Load().(tls.Certificate)
		return &cert, nil
	}
	tlsConfig := &tls.Config{
		GetCertificate:        c.Config.GetCert,
		MinVersion:            tls.VersionTLS12,
		Certificates:          []tls.Certificate{cert},
		VerifyPeerCertificate: c.Config.VerifyCertificate,
	}
	if c.Config.ServerName != "" {
		tlsConfig.ServerName = c.Config.ServerName
	}
	if len(c.Config.ServerRootCAs) > 0 {
		tlsConfig.RootCAs = x509.NewCertPool()
		for _, certBytes := range c.Config.ServerRootCAs {
			err := AddPemToCertPool(certBytes, tlsConfig.RootCAs)
			if err != nil {
				panic(err)
			}
		}
	}
	return append(opts, g.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
}
