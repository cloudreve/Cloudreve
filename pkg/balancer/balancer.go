package balancer

type Balancer interface {
	NextPeer(nodes interface{}) (error, interface{})
}

// NewBalancer 根据策略标识返回新的负载均衡器
func NewBalancer(strategy string) Balancer {
	switch strategy {
	case "RoundRobin":
		return &RoundRobin{}
	default:
		return &RoundRobin{}
	}
}
