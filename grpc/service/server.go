package service

import (
	"crypto/tls"
	"crypto/x509"
	gm "github.com/grpc-ecosystem/go-grpc-middleware"
	g "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"net"
	"sync/atomic"
	c "xianhetian.com/tendermint/p2p/grpc/config"
	i "xianhetian.com/tendermint/p2p/grpc/interceptor"
)

type Server struct {
	g.Server
	Listener    net.Listener
	Interceptor *i.Interceptor
	Config      *c.ServerConfig
}

func NewServer() *Server {
	var us []g.UnaryServerInterceptor
	return &Server{
		Config:      c.DefaultServerConfig(),
		Interceptor: &i.Interceptor{UnaryInterceptors: append(us, i.Logger)},
	}
}

func (s *Server) Listen() {
	var opts []g.ServerOption
	opts = append(opts, s.setupSecure()...)
	opts = append(opts, s.setupKeepalive()...)
	opts = append(opts, s.setupInterceptor()...)
	opts = append(opts, g.MaxSendMsgSize(s.Config.MaxSendMsgSize))
	opts = append(opts, g.MaxRecvMsgSize(s.Config.MaxRecvMsgSize))
	s.Server = *g.NewServer(opts...)
	ll, err := net.Listen("tcp", s.Config.Addr)
	if err != nil {
		panic(err)
	}
	s.Listener = ll
}

func (s *Server) setupSecure() []g.ServerOption {
	var opts []g.ServerOption
	var SecureCertificate atomic.Value
	if !s.Config.UseTLS {
		return nil
	}
	cert, err := tls.X509KeyPair(s.Config.Certificate, s.Config.Key)
	if err != nil {
		panic(err)
	}
	SecureCertificate.Store(cert)
	s.Config.GetCert = func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		cert := SecureCertificate.Load().(tls.Certificate)
		return &cert, nil
	}
	serverTLS := &tls.Config{
		GetCertificate:        s.Config.GetCert,
		CipherSuites:          s.Config.CipherSuites,
		Certificates:          []tls.Certificate{cert},
		ClientAuth:            tls.RequireAndVerifyClientCert,
		VerifyPeerCertificate: s.Config.VerifyCertificate,
	}
	if len(s.Config.ClientRootCAs) > 0 {
		serverTLS.ClientCAs = x509.NewCertPool()
		for _, ca := range s.Config.ClientRootCAs {
			certs, _, err := pemToX509Certs(ca)
			if err != nil {
				panic(err)
			}
			if len(certs) < 1 {
				panic("failed to append client root certificate(s): No client root certificates found")
			}
			for _, cert := range certs {
				serverTLS.ClientCAs.AddCert(cert)
			}
		}
	}
	return append(opts, g.Creds(credentials.NewTLS(serverTLS)))
}

func (s Server) setupKeepalive() []g.ServerOption {
	if !s.Config.UseKeepalive {
		return nil
	}
	var opts []g.ServerOption
	kap := keepalive.ServerParameters{Time: s.Config.Interval, Timeout: s.Config.AliveTimeout}
	opts = append(opts, g.KeepaliveParams(kap))
	kep := keepalive.EnforcementPolicy{MinTime: s.Config.MinInterval, PermitWithoutStream: s.Config.PermitWithoutStream}
	opts = append(opts, g.KeepaliveEnforcementPolicy(kep))
	opts = append(opts, g.ConnectionTimeout(s.Config.Timeout))
	return opts
}

func (s *Server) setupInterceptor() []g.ServerOption {
	var opts []g.ServerOption
	if len(s.Interceptor.UnaryInterceptors) > 0 {
		opts = append(opts, g.UnaryInterceptor(gm.ChainUnaryServer(s.Interceptor.UnaryInterceptors...)))
	}
	if len(s.Interceptor.StreamInterceptors) > 0 {
		opts = append(opts, g.StreamInterceptor(gm.ChainStreamServer(s.Interceptor.StreamInterceptors...)))
	}
	return opts
}
