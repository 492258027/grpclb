package main

import (
	"golang.org/x/net/context"
	"grpclb/balancer"
	"grpclb/common"
	"grpclb/consul"
	pb "grpclb/examples/proto"
	"log"
	"strconv"
	"time"
)

const (
	// SerName 服务名称
	SerName string = "route"
	SerVer  string = "1.0"
)

var (
	consulHost = "192.168.73.3"
	consulPort = 8500
	client     pb.SimpleClient
)

func main() {
	//注意 sername带上版本。  balancer 默认 "round_robin"
	conn, err := consul.InitResolver(consulHost, consulPort, SerName+":"+SerVer, balancer.RoundRobin)
	if err != nil {
		log.Fatal("Fail to register consul", err)
	}
	defer conn.Close()
	// 建立gRPC连接
	client = pb.NewSimpleClient(conn)
	for i := 0; i < 100; i++ {
		route(i)
		time.Sleep(3 * time.Second)
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
