package main

import (
	"golang.org/x/net/context"
	"grpclb/balancer"
	"grpclb/etcd3"
	pb "grpclb/examples/proto"
	"log"
	"strconv"
	"time"
)

const (
	// SerName 服务名称
	SerName string = "route"
	// SerVer 服务版本号
	SerVer string = "1.0"
)

var (
	// EtcdEndpoints etcd地址
	EtcdEndpoints = []string{"192.168.73.3:12379"}
	// SerName 服务名称
	client pb.SimpleClient
)

func main() {
	conn, err := etcd3.InitResolver(EtcdEndpoints, SerName+"/"+SerVer, balancer.RoundRobin)
	if err != nil {
		log.Fatalf("net.Connect err: %v", err)
	}
	defer conn.Close()

	// 建立gRPC连接
	client = pb.NewSimpleClient(conn)
	for i := 0; i < 100; i++ {
		route(i)
		time.Sleep(1 * time.Second)
	}
}

// route 调用服务端Route方法
func route(i int) {
	// 创建发送结构体
	req := pb.SimpleRequest{
		Data: "grpc " + strconv.Itoa(i),
	}

	// 同时传入了一个 context.Context ，在有需要时可以让我们改变RPC的行为，比如超时/取消一个正在运行的RPC
	//res, err := client.Route(context.WithValue(context.Background(), common.InstanceId, "1"), &req)
	res, err := client.Route(context.Background(), &req)
	if err != nil {
		log.Fatalf("Call Route err: %v", err)
	}
	// 打印返回值
	log.Println(res)
}
