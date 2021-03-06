package balancer

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"grpclb/common"
	"log"
	"sync"
)

const InstanceID = "instance_x"

func init() {
	balancer.Register(newInstanceBuilder())
}

func newInstanceBuilder() balancer.Builder {
	return base.NewBalancerBuilderV2(InstanceID, &instancePickerBuilder{}, base.Config{HealthCheck: true})
}

type instancePickerBuilder struct {
}

//每次地址集有改变，均调用
func (r *instancePickerBuilder) Build(buildInfo base.PickerBuildInfo) balancer.V2Picker {

	if len(buildInfo.ReadySCs) == 0 {
		return base.NewErrPickerV2(balancer.ErrNoSubConnAvailable)
	}

	picker := &instancePicker{
		subConns: make(map[string]balancer.SubConn),
	}

	for subConn, subConnInfo := range buildInfo.ReadySCs {
		if id, err := common.GetInstanceId(subConnInfo.Address); err == nil {
			picker.subConns[id] = subConn
		}
	}

	log.Println("build subConn: ", picker.subConns)

	return picker
}

type instancePicker struct {
	subConns map[string]balancer.SubConn
	mu       sync.Mutex
}

func (p *instancePicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	ret := balancer.PickResult{}

	p.mu.Lock()
	if ins := info.Ctx.Value(common.InstanceId); ins != nil{
		if m, ok := ins.(string); ok {
			if v, ok := p.subConns[m]; ok {
				ret.SubConn = v
			}
		}
	}
	log.Println("pick SubConn: ", ret.SubConn)
	p.mu.Unlock()

	return ret, nil
}
