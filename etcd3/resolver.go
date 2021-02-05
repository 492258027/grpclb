package etcd3

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"grpclb/common"
	"log"
	"sync"
	"time"
)

const schema = "juzhouyun"

type etcdResolver struct {
	cc         resolver.ClientConn
	cli        *clientv3.Client
	serverList sync.Map
}

func NewEtcdResolver(ep []string) resolver.Builder {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   ep,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}

	return &etcdResolver{
		cli: cli,
	}
}

//当调用grpc.Dial()时执行一次, 根据dail调用时传入的第一个参数target解出相关查询条件，并返回一个resolver。
func (r *etcdResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	log.Println("Build, only once!!!")

	r.cc = cc
	prefix := "/" + target.Scheme + "/" + target.Endpoint + "/"

	//根据前缀获取现有的key并更新state
	resp, err := r.cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, ev := range resp.Kvs {
		r.setServiceList(string(ev.Key), string(ev.Value))
	}
	r.cc.UpdateState(resolver.State{Addresses: r.getAllServices()})

	//监控变更
	go r.watcher(prefix)

	return r, nil
}

//Scheme return schema
func (r *etcdResolver) Scheme() string {
	return schema
}

// ResolveNow 监视目标更新
func (r *etcdResolver) ResolveNow(rn resolver.ResolveNowOption) {
	log.Println("ResolveNow")
}

//Close 关闭
func (r *etcdResolver) Close() {
	log.Println("Close")
	r.cli.Close()
}

//watcher 监听前缀
func (r *etcdResolver) watcher(prefix string) {
	rch := r.cli.Watch(context.Background(), prefix, clientv3.WithPrefix())
	log.Printf("watching prefix:%s now...", prefix)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			switch ev.Type {
			case mvccpb.PUT: //新增或修改
				r.setServiceList(string(ev.Kv.Key), string(ev.Kv.Value))
			case mvccpb.DELETE: //删除
				r.delServiceList(string(ev.Kv.Key))
			}
		}
	}
}

func (r *etcdResolver) setServiceList(key, val string) {
	//从etcd中读取json字符串并解析为自定义数据结构
	var v common.ServiceInfo
	err := json.Unmarshal([]byte(val), &v)
	if err != nil {
		log.Println("unmarshal error:", err)
		return
	}

	att := attributes.New()
	//组weight
	if v, ok := v.Metadata[common.WeightKey]; ok {
		att = att.WithValues(common.WeightKey, v)
	}
	//组instanceId
	if v, ok := v.Metadata[common.InstanceId]; ok {
		att = att.WithValues(common.InstanceId, v)
	}

	//把自定义数据结构转换成resolver.Address， 并写入map表
	r.serverList.Store(key, resolver.Address{
		Addr:       fmt.Sprintf("%s:%d", v.Ip, v.Port),
		Attributes: att},
	)

	//从map表中读取所有的resolver.Address，并组成切片，赋值给resolver
	r.cc.UpdateState(resolver.State{Addresses: r.getAllServices()})
	log.Println("put key :", key, "val:", val)
}

func (r *etcdResolver) delServiceList(key string) {
	r.serverList.Delete(key)
	r.cc.UpdateState(resolver.State{Addresses: r.getAllServices()})
	log.Println("del key:", key)
}

//遍历map并返回所有value
func (s *etcdResolver) getAllServices() []resolver.Address {
	addrs := make([]resolver.Address, 0, 10)
	s.serverList.Range(func(k, v interface{}) bool {
		addrs = append(addrs, v.(resolver.Address))
		return true
	})
	return addrs
}

//封装一下， 方便用户调用, 用不用均可
func InitResolver(ep []string, serName, balbancerName string) (*grpc.ClientConn, error) {

	r := NewEtcdResolver(ep)
	resolver.Register(r)

	conn, err := grpc.Dial(
		fmt.Sprintf("%s:///%s", r.Scheme(), serName),
		grpc.WithBalancerName(balbancerName),
		grpc.WithInsecure(),
	)
	if err != nil {
		log.Println("net.Connect err", err)
		return nil, err
	}

	return conn, nil
}
