package config

import (
	"crypto/tls"
	"crypto/x509"
	c "google.golang.org/grpc/credentials"
	"time"
)

// Keepalive配置选项
type KeepaliveOptions struct {
	UseKeepalive        bool          `mapstructure:"use_keepalive"`         // 是否使用Keepalive
	Interval            time.Duration `mapstructure:"interval"`              // ping间隔时间
	MinInterval         time.Duration `mapstructure:"min_interval"`          // ping最短间隔时间
	AliveTimeout        time.Duration `mapstructure:"alive_timeout"`         // 发出ping之后等待超时时间
	PermitWithoutStream bool          `mapstructure:"permit_without_stream"` // 是否在没有GPRC活动流的情况下继续保持需要接收ping
}

// TLS配置选项
type SecureOptions struct {
	UseTLS             bool                                                                // 是否使用TLS
	InsecureSkipVerify bool                                                                // 是否验证服务器的证书链和主机名
	Key                []byte                                                              // PEM编码私钥
	Certificate        []byte                                                              // PEM编码X509公钥
	CipherSuites       []uint16                                                            // TLS密钥算法套件
	PreCreds           c.PerRPCCredentials                                                 // 自定义认证（如Token）
	ServerRootCAs      [][]byte                                                            // 客户端用于验证服务器证书的PEM编码的X509证书颁发机构的集合
	ClientRootCAs      [][]byte                                                            // 服务器用于验证客户端证书的PEM编码的X509证书颁发机构的集合
	GetCert            func(*tls.ClientHelloInfo) (*tls.Certificate, error)                // TLS证书原子存储
	VerifyCertificate  func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error // 自定义验证证书方法
}
