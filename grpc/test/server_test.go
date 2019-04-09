package test

import (
	"context"
	"io"
	"strconv"
	"strings"
	"sync"
	"testing"
	"xianhetian.com/tendermint/p2p/grpc/service"
)

func TestServer(t *testing.T) {
	// 1.封装参数
	s := service.NewServer()
	s.Config.UseTLS = false
	// 2.启动服务
	s.Listen()
	RegisterGreeterServer(&s.Server, &greeterServer{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		s.Serve(s.Listener)
		wg.Done()
	}()
	wg.Wait()
	s.Stop()
}

type greeterServer struct{}

func (h greeterServer) SayHello(ctx context.Context, in *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{Message: "Hello " + in.Name}, nil
}

func (h greeterServer) LotsOfReplies(in *HelloRequest, gl Greeter_LotsOfRepliesServer) error {
	for i := 0; i < 100; i++ {
		gl.Send(&HelloResponse{Message: "Hello " + in.Name + strconv.Itoa(i)})
	}
	return nil
}

func (h greeterServer) LotsOfGreetings(gl Greeter_LotsOfGreetingsServer) error {
	var names []string
	for {
		in, err := gl.Recv()
		// 接受信息完毕
		if err == io.EOF {
			gl.SendAndClose(&HelloResponse{Message: "Hello " + strings.Join(names, ",")})
			return nil
		}
		names = append(names, in.Name)
	}
	return nil
}

func (h greeterServer) BidiHello(gb Greeter_BidiHelloServer) error {
	for {
		in, err := gb.Recv()
		// 接受信息完毕
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		gb.Send(&HelloResponse{Message: "Hello " + in.Name})
	}
	return nil
}
