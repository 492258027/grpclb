package etcd3

import (
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"golang.org/x/net/context"
	"grpclb/common"
	"log"
	"time"
)

type etcdRegister struct {
	cli     *clientv3.Client
	leaseID clientv3.LeaseID //租约ID
}

func NewEtcdRegister(ep []string) (*etcdRegister, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   ep,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &etcdRegister{
		cli: client,
	}, nil
}

func (r *etcdRegister) Register(service *common.ServiceInfo, lease int64) error {
	val, err := json.Marshal(service)
	if err != nil {
		return err
	}

	key := "/" + schema + "/" + service.SerName + "/" + service.Version + "/" + service.InstanceId
	value := string(val)

	//设置租约时间
	resp, err := r.cli.Grant(context.Background(), lease)
	if err != nil {
		return err
	}
	//注册服务并绑定租约
	_, err = r.cli.Put(context.Background(), key, value, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	//设置续租 定期发送需求请求
	leaseRespChan, err := r.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return err
	}
	log.Printf("Put key:%s  val:%s  success!", key, value)

	//保存leaseID，注销时用
	r.leaseID = resp.ID

	go func() {
		for {
			<-leaseRespChan
		}
	}()

	return nil
}

func (r *etcdRegister) Unregister() error {
	//撤销租约
	if _, err := r.cli.Revoke(context.Background(), r.leaseID); err != nil {
		return err
	}

	log.Println("撤销租约")
	return r.cli.Close()
}

//封装一下， 方便用户调用, 用不用均可
func InitRegister(ep []string, service *common.ServiceInfo, lease int64) (*etcdRegister, error) {
	r, err := NewEtcdRegister(ep)
	if err != nil {
		log.Println("Fail to new client", err)
		return nil, err
	}

	if err := r.Register(service, lease); err != nil {
		log.Println("Fail to register service", err)
		return nil, err
	}

	return r, nil
}
