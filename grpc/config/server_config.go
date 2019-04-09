package config

import (
	"crypto/tls"
	"io/ioutil"
	"time"
)

var (
	DefaultTLSCipherSuites = []uint16{
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	}
)

// 服务端配置
type ServerConfig struct {
	BaseServerConfig `mapstructure:"base_server_Config"`
	SecureOptions    `mapstructure:"secure_options"`
	KeepaliveOptions `mapstructure:"keepalive_options"`
}

// 基础配置
type BaseServerConfig struct {
	Addr           string        `mapstructure:"addr"`              // 监听地址
	CaPath         string        `mapstructure:"ca_path"`           // CA文件路径
	KeyPath        string        `mapstructure:"key_path"`          // 私钥文件路径
	CertPath       string        `mapstructure:"cert_path"`         // 证书文件路径
	Timeout        time.Duration `mapstructure:"timeout"`           // 新连接建立超时时间
	MaxRecvMsgSize int           `mapstructure:"max_recv_msg_size"` // 服务端可以接收的最大信息字节数
	MaxSendMsgSize int           `mapstructure:"max_send_msg_size"` // 服务端可以发送的最大信息字节数
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		BaseServerConfig: *DefaultBaseServerConfig(),
		SecureOptions:    *DefaultServerSecureOptions(),
		KeepaliveOptions: *DefaultServerKeepaliveOptions(),
	}
}

func DefaultBaseServerConfig() *BaseServerConfig {
	return &BaseServerConfig{
		Addr:           "0.0.0.0:36658",
		CaPath:         "/files/sources/src/xianhetian.com/tendermint/p2p/grpc/config/ca.pem",
		KeyPath:        "/files/sources/src/xianhetian.com/tendermint/p2p/grpc/config/server.key",
		CertPath:       "/files/sources/src/xianhetian.com/tendermint/p2p/grpc/config/server.pem",
		MaxRecvMsgSize: 100 * 1024 * 1024,
		MaxSendMsgSize: 100 * 1024 * 1024,
		Timeout:        5 * time.Second,
	}
}

func DefaultServerKeepaliveOptions() *KeepaliveOptions {
	return &KeepaliveOptions{
		UseKeepalive:        true,
		Interval:            time.Duration(2) * time.Hour,
		MinInterval:         time.Duration(1) * time.Minute,
		AliveTimeout:        time.Duration(20) * time.Second,
		PermitWithoutStream: true,
	}
}

func DefaultServerSecureOptions() *SecureOptions {
	ca, _ := ioutil.ReadFile(DefaultBaseServerConfig().CaPath)
	key, _ := ioutil.ReadFile(DefaultBaseServerConfig().KeyPath)
	cert, _ := ioutil.ReadFile(DefaultBaseServerConfig().CertPath)
	return &SecureOptions{
		UseTLS:        true,
		Key:           key,
		Certificate:   cert,
		ClientRootCAs: [][]byte{ca},
		CipherSuites:  DefaultTLSCipherSuites,
	}
}
