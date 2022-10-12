package direct

import (
	"context"
	wmeta "github.com/pll/kratos/pkg/naming/metadata"
	"github.com/pll/kratos/pkg/net/rpc/warden/balancer/p2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/serviceconfig"
	"strings"

	"github.com/pll/kratos/pkg/naming"
	gresolver "google.golang.org/grpc/resolver"
)

const (
	// Name is the name of direct resolver
	Name = "direct"
)

var _ naming.Resolver = &Direct{}

// New return Direct
func New() *Direct {
	return &Direct{}
}

// Build build direct.
func Build(id string) *Direct {
	return &Direct{id: id}
}

// Direct is a resolver for conneting endpoints directly.
// example format: direct://default/192.168.1.1:8080,192.168.1.2:8081
type Direct struct {
	id string
}

// Build direct build.
func (d *Direct) Build(id string, opt ...naming.BuildOpt) naming.Resolver {
	return &Direct{id: id}
}

// Scheme return the Scheme of Direct
func (d *Direct) Scheme() string {
	return Name
}

// Watch a tree.
func (d *Direct) Watch() <-chan struct{} {
	ch := make(chan struct{}, 1)
	ch <- struct{}{}
	return ch
}

// Unwatch a tree.
func (d *Direct) Unwatch(id string) {
}

//Fetch fetch isntances.
func (d *Direct) Fetch(ctx context.Context) (State gresolver.State, ok bool) {
	var weight int64
	weight = 10

	addrs := strings.Split(d.id, ",")
	stateaddrs := make([]gresolver.Address, 0, len(addrs))
	for _, addr := range addrs {
		//ins = append(ins, &naming.Instance{
		//	Addrs: []string{fmt.Sprintf("%s://%s", resolver.Scheme, addr)},
		//})
		stateaddr := gresolver.Address{
			Addr:     addr,
			Type:     gresolver.Backend,
			Metadata: wmeta.MD{Weight: uint64(weight)},
		}

		stateaddrs = append(stateaddrs, stateaddr)
	}

	if len(stateaddrs) > 0 {
		ok = true
	}

	//res := &naming.InstancesInfo{
	//	Instances: map[string][]*naming.Instance{env.Zone: ins},
	//}

	LB := p2c.Name
	State = gresolver.State{
		Addresses: stateaddrs,
		//默认的serviceconfig, 使用dailoption会覆盖默认config
		ServiceConfig: &serviceconfig.ParseResult{
			Config: &grpc.ServiceConfig{LB: &LB},
			Err:    nil,
		},
	}

	return
}

//Close close Direct
func (d *Direct) Close() error {
	return nil
}
