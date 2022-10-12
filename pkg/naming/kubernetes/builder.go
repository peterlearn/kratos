package kubernetes

import (
	"context"
	"fmt"
	"github.com/peterlearn/kratos/pkg/naming"
	"google.golang.org/grpc/resolver"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	KubernetesSchema = "kubernetes"
	defaultFreq      = time.Minute * 1
)

func Builder() naming.Builder {
	var k8sclient K8sClient

	if cl, err := NewInClusterK8sClient(); err == nil {
		k8sclient = cl
	} else {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	k := &K8S{
		ctx:       ctx,
		cancel:    cancel,
		k8sClient: k8sclient,
		t:         time.NewTimer(defaultFreq),
		freq:      defaultFreq,
	}

	return k

	//return &K8S{}
}

func parseResolverTarget(tistr string) (targetInfo, error) {
	target := resolver.Target{}

	str := strings.SplitN(tistr, ",", 3)

	if len(str) == 3 {
		target.Scheme = str[0]
		target.Authority = str[1]
		target.Endpoint = str[2]
	} else {
		return targetInfo{}, fmt.Errorf("target string(%s) is invalid", tistr)
	}

	// kubernetes://default/service:port
	end := target.Endpoint
	snamespace := target.Authority
	// kubernetes://service.default:port/
	if end == "" {
		end = target.Authority
		snamespace = ""
	}
	ti := targetInfo{}
	if end == "" {
		return targetInfo{}, fmt.Errorf("target(%q) is empty", target)
	}
	var name string
	var port string
	if strings.LastIndex(end, ":") < 0 {
		name = end
		port = ""
		ti.useFirstPort = true
	} else {
		var err error
		name, port, err = net.SplitHostPort(end)
		if err != nil {
			return targetInfo{}, fmt.Errorf("target endpoint='%s' is invalid. grpc target is %#v, err=%v", end, target, err)
		}
	}

	namesplit := strings.SplitN(name, ".", 2)
	sname := name
	podname := ""
	if len(namesplit) == 2 {
		sname = namesplit[0]
		snamespace = namesplit[1]
	}
	podsplit := strings.SplitN(sname, "/", 2)
	if len(podsplit) == 2 {
		sname = podsplit[0]
		podname = podsplit[1]
	}
	ti.serviceName = sname
	ti.serviceNamespace = snamespace
	ti.podname = podname
	ti.port = port
	if !ti.useFirstPort {
		if _, err := strconv.Atoi(ti.port); err != nil {
			ti.resolveByPortName = true
		} else {
			ti.resolveByPortName = false
		}
	}
	return ti, nil
}
