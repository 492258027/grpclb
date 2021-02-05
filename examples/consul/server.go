package main

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"grpclb/common"
	"grpclb/consul"
	pb "grpclb/examples/proto"
	"log"
	"net"
)

const (
	// grpc服务监听地址
	SerIp string = "10.17.9.246"
	// grpc服务监听端口
	SerPort int = 8000
	// SerName 服务名称
	SerName string = "route"
	// SerVer 服务版本号
	SerVer string = "1.0"
)

var (
	consulHost = "192.168.73.3"
	consulPort = 8500
)

func main() {
	// 监听本地端口
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", SerIp, SerPort))
	if err != nil {
		log.Fatalf("net.Listen err: %v", err)
	}

	// 在gRPC服务器注册我们的服务
	grpcServer := grpc.NewServer()
	pb.RegisterSimpleServer(grpcServer, &SimpleService{})
	//consul的健康检查
	grpc_health_v1.RegisterHealthServer(grpcServer, &HealthImpl{})

	id := uuid.NewV4().String()
	r, err := consul.InitRegister(consulHost, consulPort, &common.ServiceInfo{
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
		log.Fatal("Fail to register consul", err)
	}
	defer r.Unregister()

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatalln(err)
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

// HealthImpl 健康检查实现
type HealthImpl struct{}

// Check 实现健康检查接口，这里直接返回健康状态，这里也可以有更复杂的健康检查策略，比如根据服务器负载来返回
func (h *HealthImpl) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	//log.Println("health checking\n")
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthImpl) Watch(req *grpc_health_v1.HealthCheckRequest, w grpc_health_v1.Health_WatchServer) error {
	return nil
}
