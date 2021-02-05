package main

import (
	"context"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"grpclb/common"
	"grpclb/etcd3"
	pb "grpclb/examples/proto"
	"log"
	"net"
)

const (
	// SerIp 服务监听地址
	SerIp string = "10.17.9.246"
	// SerPort 服务监听端口
	SerPort int = 8000
	// SerName 服务名称
	SerName string = "route"
	// SerVer 服务版本号
	SerVer string = "1.0"
)

var etcdEndpoint = []string{"192.168.73.3:12379"}

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", SerIp, SerPort))
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}

	// 在gRPC服务器注册我们的服务
	grpcServer := grpc.NewServer()
	pb.RegisterSimpleServer(grpcServer, &SimpleService{})

	id := uuid.NewV4().String()
	r, err := etcd3.InitRegister(etcdEndpoint, &common.ServiceInfo{
		InstanceId: id,
		SerName:    SerName,
		Version:    SerVer,
		Ip:         SerIp,
		Port:       SerPort,
		Metadata: map[string]string{
			common.WeightKey:  "1",
			common.InstanceId: id,
		},
	}, 5)
	if err != nil {
		log.Fatalf("register service err: %v", err)
	}
	defer r.Unregister()

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalf("grpcServer.Serve err: %v", err)
	}
}

type SimpleService struct{}

// Route 实现Route方法
func (s *SimpleService) Route(ctx context.Context, req *pb.SimpleRequest) (*pb.SimpleResponse, error) {
	log.Println("receive: " + req.Data)
	res := pb.SimpleResponse{
		Code:  200,
		Value: "hello " + req.Data,
	}
	return &res, nil
}
