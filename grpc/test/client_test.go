package test

import (
	"context"
	"github.com/stretchr/testify/require"
	"io"
	"strconv"
	"testing"
	"xianhetian.com/tendermint/p2p/grpc/service"
)

// 简单rpc，一个请求对象对应一个返回对象
func TestResponse(t *testing.T) {
	// 1.封装参数
	cp := service.NewClient()
	cp.Config.UseTLS = false
	// 2.根据不同的地址创建连接
	_, err := cp.Conn("0.0.0.0:36658")
	if err != nil {
		panic(err)
	}
	defer cp.Close()
	// 3.初始化客户端调用方法
	response, err := NewGreeterClient(&cp.ClientConn).SayHello(context.Background(), &HelloRequest{Name: "World"})
	require := require.New(t)
	require.Nil(err, "%+v", err)
	// 4.验证返回结果是否正确
	require.EqualValues("Hello World", response.Message)
}

// 服务端流式，一个请求对象，服务端可以传回多个结果对象
func TestRepliesClient(t *testing.T) {
	// 1.封装参数
	cp := service.NewClient()
	cp.Config.UseTLS = false
	// 2.根据不同的地址创建连接
	cp.Conn()
	defer cp.Close()
	// 3.初始化客户端调用方法
	repliesClient, err := NewGreeterClient(&cp.ClientConn).LotsOfReplies(context.Background(), &HelloRequest{Name: "num"})
	require := require.New(t)
	require.Nil(err, "%+v", err)
	var i int64
	for {
		reply, err := repliesClient.Recv()
		if err == io.EOF {
			break
		}
		// 4.验证返回结果是否正确
		require.EqualValues("Hello num"+strconv.FormatInt(i, 10), reply.Message)
		i++
	}
}

// 客户端流式，客户端传入多个请求对象，服务端返回一个响应结果
func TestGreetingsClient(t *testing.T) {
	// 1.封装参数
	cp := service.NewClient()
	cp.Config.UseTLS = false
	// 2.根据不同的地址创建连接
	cp.Conn()
	defer cp.Close()
	// 3.初始化流式客户端
	greetingsClient, err := NewGreeterClient(&cp.ClientConn).LotsOfGreetings(context.Background())
	require := require.New(t)
	require.Nil(err, "%+v", err)
	for i := 0; i < 100; i++ {
		greetingsClient.Send(&HelloRequest{Name: "num" + strconv.Itoa(i)})
	}
	// 4.输出返回结果
	reply, err := greetingsClient.CloseAndRecv()
	require.Nil(err, "%+v", err)
	t.Logf("Greeting:%s", reply.Message)
}

// 双向流式，可以传入多个对象，返回多个响应对象，类似TCP的client和server
func TestBidiHelloClient(t *testing.T) {
	// 1.封装参数
	cp := service.NewClient()
	cp.Config.UseTLS = false
	// 2.根据不同的地址创建连接
	cp.Conn()
	defer cp.Close()
	// 3.初始化流式客户端
	bidiHelloClient, err := NewGreeterClient(&cp.ClientConn).BidiHello(context.Background())
	require := require.New(t)
	require.Nil(err, "%+v", err)
	var i int64
	for {
		bidiHelloClient.Send(&HelloRequest{Name: "num" + strconv.FormatInt(i, 10)})
		reply, err := bidiHelloClient.Recv()
		require.Nil(err, "%+v", err)
		if i == 2 {
			// 4.验证返回结果是否正确
			require.EqualValues("Hello num"+strconv.FormatInt(i, 10), reply.Message)
			break
		}
		i++
	}
	bidiHelloClient.CloseSend()
}
