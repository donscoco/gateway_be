package load_balance

type LbType int

const (
	LbRandom LbType = iota
	LbRoundRobin
	LbWeightRoundRobin
	LbConsistentHash
)

func LoadBalanceFactory(lbType LbType) LoadBalance {
	switch lbType {
	case LbRandom:
		return &RandomBalance{}
	case LbConsistentHash:
		return NewConsistentHashBanlance(10, nil)
	case LbRoundRobin:
		return &RoundRobinBalance{}
	case LbWeightRoundRobin:
		return &WeightRoundRobinBalance{}
	default:
		return &RandomBalance{}
	}
}

func LoadBalanceFactorWithConf(lbType LbType, mConf LoadBalanceConf) LoadBalance {
	//观察者模式
	switch lbType {
	case LbRandom:
		lb := &RandomBalance{}
		lb.SetConf(mConf)
		mConf.Attach(lb) // 将lb 设置到config的观察者中，一旦改动就通知这些 lb
		lb.Update()
		return lb
	case LbConsistentHash:
		lb := NewConsistentHashBanlance(10, nil)
		lb.SetConf(mConf)
		mConf.Attach(lb)
		lb.Update()
		return lb
	case LbRoundRobin:
		lb := &RoundRobinBalance{}
		lb.SetConf(mConf)
		mConf.Attach(lb)
		lb.Update()
		return lb
	case LbWeightRoundRobin:
		lb := &WeightRoundRobinBalance{}
		lb.SetConf(mConf)
		mConf.Attach(lb)
		lb.Update()
		return lb
	default:
		lb := &RandomBalance{}
		lb.SetConf(mConf)
		mConf.Attach(lb)
		lb.Update()
		return lb
	}
}
