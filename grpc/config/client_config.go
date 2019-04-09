package config

import (
	"context"
	c "google.golang.org/grpc/credentials"
	"io/ioutil"
	"time"
)

// 客户端配置
type ClientConfig struct {
	BaseClientConfig `mapstructure:"base_client_Config"`
	SecureOptions    `mapstructure:"secure_options"`
	KeepaliveOptions `mapstructure:"keepalive_options"`
}

// 基础配置
type BaseClientConfig struct {
	Addr           string        `mapstructure:"addr"`              // 服务端地址
	CaPath         string        `mapstructure:"ca_path"`           // CA文件路径
	KeyPath        string        `mapstructure:"key_path"`          // 私钥文件路径
	CertPath       string        `mapstructure:"cert_path"`         // 证书文件路径
	ServerName     string        `mapstructure:"server_name"`       // 服务端虚拟host名称
	Timeout        time.Duration `mapstructure:"timeout"`           // 客户端在尝试建立连接时阻塞的超时时间
	MaxRecvMsgSize int           `mapstructure:"max_recv_msg_size"` // 客戶端可以接收的最大信息字节数
	MaxSendMsgSize int           `mapstructure:"max_send_msg_size"` // 客戶端可以发送的最大信息字节数
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		BaseClientConfig: *DefaultBaseClientConfig(),
		SecureOptions:    *DefaultClientSecureOptions(),
		KeepaliveOptions: *DefaultClientKeepaliveOptions(),
	}
}

func DefaultBaseClientConfig() *BaseClientConfig {
	return &BaseClientConfig{
		Addr:           "0.0.0.0:36658",
		ServerName:     "go-grpc-example",
		CaPath:         "/files/sources/src/xianhetian.com/tendermint/p2p/grpc/config/ca.pem",
		KeyPath:        "/files/sources/src/xianhetian.com/tendermint/p2p/grpc/config/client.key",
		CertPath:       "/files/sources/src/xianhetian.com/tendermint/p2p/grpc/config/client.pem",
		Timeout:        3 * time.Second,
		MaxRecvMsgSize: 100 * 1024 * 1024,
		MaxSendMsgSize: 100 * 1024 * 1024,
	}
}

func DefaultClientSecureOptions() *SecureOptions {
	ca, _ := ioutil.ReadFile(DefaultBaseServerConfig().CaPath)
	key, _ := ioutil.ReadFile(DefaultBaseServerConfig().KeyPath)
	cert, _ := ioutil.ReadFile(DefaultBaseServerConfig().CertPath)
	return &SecureOptions{
		UseTLS:        true,
		Key:           key,
		Certificate:   cert,
		ServerRootCAs: [][]byte{ca},
	}
}

func DefaultClientKeepaliveOptions() *KeepaliveOptions {
	return &KeepaliveOptions{
		UseKeepalive: true,
		Interval:     time.Duration(1) * time.Minute,
		AliveTimeout: time.Duration(20) * time.Second,
	}
}

type Token struct {
	token string
}

func GetToken(token string) c.PerRPCCredentials {
	return Token{token: string(token)}
}

func (t Token) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{"authorization": t.token}, nil
}

func (t Token) RequireTransportSecurity() bool {
	return true
}
