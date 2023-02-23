package service_manager

import (
	"fmt"
	"github.com/donscoco/gateway_be/internal/bl"
	"github.com/donscoco/gateway_be/internal/dao"
	"github.com/donscoco/gateway_be/internal/reverse_proxy/load_balance"
	"sync"
)

var LoadBalancerHandler *LoadBalancer

type LoadBalancer struct {
	LoadBanlanceMap   map[string]*LoadBalancerItem
	LoadBanlanceSlice []*LoadBalancerItem
	Locker            sync.RWMutex
}

type LoadBalancerItem struct {
	LoadBanlance load_balance.LoadBalance
	ServiceName  string
}

func NewLoadBalancer() *LoadBalancer {
	return &LoadBalancer{
		LoadBanlanceMap:   map[string]*LoadBalancerItem{},
		LoadBanlanceSlice: []*LoadBalancerItem{},
		Locker:            sync.RWMutex{},
	}
}

func init() {
	LoadBalancerHandler = NewLoadBalancer()
}

func (lbr *LoadBalancer) GetLoadBalancer(service *dao.ServiceDetail) (load_balance.LoadBalance, error) {
	for _, lbrItem := range lbr.LoadBanlanceSlice {
		if lbrItem.ServiceName == service.Info.ServiceName {
			return lbrItem.LoadBanlance, nil
		}
	}
	schema := "http://"
	if service.HTTPRule.NeedHttps == 1 {
		schema = "https://"
	}
	if service.Info.LoadType == bl.LoadTypeTCP || service.Info.LoadType == bl.LoadTypeGRPC {
		schema = ""
	}
	ipList := service.LoadBalance.GetIPListByModel()
	weightList := service.LoadBalance.GetWeightListByModel()
	ipConf := map[string]string{}
	for ipIndex, ipItem := range ipList {
		ipConf[ipItem] = weightList[ipIndex]
	}
	//fmt.Println("ipConf", ipConf)
	mConf, err := load_balance.NewLoadBalanceCheckConf(fmt.Sprintf("%s%s", schema, "%s"), ipConf)
	if err != nil {
		return nil, err
	}
	lb := load_balance.LoadBanlanceFactorWithConf(load_balance.LbType(service.LoadBalance.RoundType), mConf)

	//save to map and slice
	lbItem := &LoadBalancerItem{
		LoadBanlance: lb,
		ServiceName:  service.Info.ServiceName,
	}
	lbr.LoadBanlanceSlice = append(lbr.LoadBanlanceSlice, lbItem)

	lbr.Locker.Lock()
	defer lbr.Locker.Unlock()
	lbr.LoadBanlanceMap[service.Info.ServiceName] = lbItem
	return lb, nil
}
