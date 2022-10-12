package kubernetes

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/peterlearn/kratos/pkg/log"
	"github.com/peterlearn/kratos/pkg/naming"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	serviceAccountToken     = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	serviceAccountCACert    = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	kubernetesNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	defaultNamespace        = "default"
)

type targetInfo struct {
	serviceName       string
	serviceNamespace  string
	port              string
	podname           string
	resolveByPortName bool
	useFirstPort      bool
}

type K8sClient interface {
	Do(req *http.Request) (*http.Response, error)
	GetRequest(url string) (*http.Request, error)
	Host() string
}

type k8sClient struct {
	host       string
	token      string
	httpClient *http.Client
}

func (kc *k8sClient) GetRequest(url string) (*http.Request, error) {
	if !strings.HasPrefix(url, kc.host) {
		url = fmt.Sprintf("%s/%s", kc.host, url)
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	if len(kc.token) > 0 {
		req.Header.Set("Authorization", "Bearer "+kc.token)
	}
	return req, nil
}

func (kc *k8sClient) Do(req *http.Request) (*http.Response, error) {
	return kc.httpClient.Do(req)
}

func (kc *k8sClient) Host() string {
	return kc.host
}

type K8S struct {
	ctx       context.Context
	cancel    context.CancelFunc
	k8sClient K8sClient
	endpoints atomic.Value
	// wg is used to enforce Close() to return after the watcher() goroutine has finished.
	wg   sync.WaitGroup
	t    *time.Timer
	freq time.Duration
}

func (k *K8S) Build(target string, options ...naming.BuildOpt) naming.Resolver {

	str := strings.SplitN(target, ",", 3)

	ti := resolver.Target{}
	//fmt.Println(str)
	if len(str) == 3 {
		ti.Scheme = str[0]
		ti.Authority = str[1]
		ti.Endpoint = str[2]
	} else {
		log.Error("target string error")
	}

	targetInfos, err := parseResolverTarget(target)
	if err != nil {
		return nil
	}
	if targetInfos.serviceNamespace == "" {
		targetInfos.serviceNamespace = getCurrentNamespaceOrDefault()
	}

	r := &K8sResolver{
		event:  make(chan struct{}, 1),
		k8s:    k,
		target: targetInfos,
	}

	r.resolve()

	go r.resolveproc()
	return r
}

func (k *K8S) Scheme() string {
	return KubernetesSchema
}

type K8sResolver struct {
	event  chan struct{}
	k8s    *K8S
	target targetInfo
}

func (r *K8sResolver) Watch() <-chan struct{} {
	return r.event
}

func (r *K8sResolver) Fetch(ctx context.Context) (State resolver.State, ok bool) {
	//log.Info("fetch new GRPC State")
	e, ok := r.k8s.endpoints.Load().(Endpoints)

	if !ok {
		log.Error("kuberesolver: lookup endpoints failed")
		return
	}

	State = r.makeState(e)

	return
}

func (r *K8sResolver) Close() error {
	//r.k8s.cancel()
	//r.k8s.wg.Wait()
	log.Error("GRPC ClientConn Close() try Close K8sResolver;Skip close K8sResolver")
	return nil
}

func (r *K8sResolver) Watchk8sAPI() error {
	//defer r.k8s.wg.Done()
	// watch endpoints lists existing endpoints at start
	sw, err := watchEndpoints(r.k8s.k8sClient, r.target.serviceNamespace, r.target.serviceName)
	if err != nil {
		return err
	}
	for {
		select {
		case <-r.k8s.ctx.Done():
			return nil
		case <-r.k8s.t.C:
			r.resolve()
		case up, hasMore := <-sw.ResultChan():
			if hasMore {
				r.handle(up.Object)
			} else {
				return nil
			}
		}
	}
}

func (r *K8sResolver) resolve() {
	//log.Info("resolve request k8s API get endpoints")
	e, err := getEndpoints(r.k8s.k8sClient, r.target.serviceNamespace, r.target.serviceName)
	if err == nil {
		r.handle(e)
	} else {
		log.Error("kuberesolver: lookup endpoints failed: %v", err)
	}
	// Next lookup should happen after an interval defined by k.freq.
	r.k8s.t.Reset(r.k8s.freq)
}

func (r *K8sResolver) handle(e Endpoints) {
	r.k8s.endpoints.Store(e)
	//log.Info("handle write endpoints")
	select {
	case r.event <- struct{}{}:
	default:
	}
}

func (r *K8sResolver) resolveproc() {
	watchfunc := func() {
		//r.k8s.wg.Add(1)
		err := r.Watchk8sAPI()
		if err != nil && err != io.EOF {
			log.Error("kuberesolver: watching ended with error='%v', will reconnect again", err)
		}
	}

	stopCh := r.k8s.ctx.Done()

	select {
	case <-stopCh:
		return
	default:
	}
	for {
		func() {
			defer handleCrash()
			watchfunc()
		}()
		select {
		case <-stopCh:
			return
		case <-time.After(time.Second):

		}
	}
}

func handleCrash() {
	if r := recover(); r != nil {
		callers := string(debug.Stack())
		log.Error("kuberesolver: recovered from panic: %#v (%v)\n%v", r, r, callers)
	}
}

func (r *K8sResolver) makeState(e Endpoints) resolver.State {
	var newAddrs []resolver.Address
	for _, subset := range e.Subsets {
		port := ""
		if r.target.useFirstPort {
			port = strconv.Itoa(subset.Ports[0].Port)
		} else if r.target.resolveByPortName {
			for _, p := range subset.Ports {
				if p.Name == r.target.port {
					port = strconv.Itoa(p.Port)
					break
				}
			}
		} else {
			port = r.target.port
		}

		if len(port) == 0 {
			port = strconv.Itoa(subset.Ports[0].Port)
		}

		for _, address := range subset.Addresses {
			sname := r.target.serviceName
			if address.TargetRef != nil {
				//只取出podname=TargetRef.Name的服务地址
				if r.target.podname != "" && r.target.podname != address.TargetRef.Name {
					continue
				}

				sname = address.TargetRef.Name
			}
			newAddrs = append(newAddrs, resolver.Address{
				Type:       resolver.Backend,
				Addr:       net.JoinHostPort(address.IP, port),
				ServerName: sname,
				Metadata:   nil,
			})
		}
	}

	//k8s默认只支持RR/FP负载均衡
	//可使用dailoption覆盖默认config
	LB := "round_robin"
	State := resolver.State{
		Addresses: newAddrs,
		//默认的serviceconfig, 使用dailoption会覆盖默认config
		ServiceConfig: &serviceconfig.ParseResult{
			Config: &grpc.ServiceConfig{LB: &LB},
			Err:    nil,
		},
	}

	return State
}

// NewInClusterK8sClient creates K8sClient if it is inside Kubernetes
func NewInClusterK8sClient() (K8sClient, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return nil, fmt.Errorf("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
	}
	token, err := ioutil.ReadFile(serviceAccountToken)
	if err != nil {
		return nil, err
	}
	ca, err := ioutil.ReadFile(serviceAccountCACert)
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(ca)
	transport := &http.Transport{TLSClientConfig: &tls.Config{
		MinVersion: tls.VersionTLS10,
		RootCAs:    certPool,
	}}
	httpClient := &http.Client{Transport: transport, Timeout: time.Nanosecond * 0}

	return &k8sClient{
		host:       "https://" + net.JoinHostPort(host, port),
		token:      string(token),
		httpClient: httpClient,
	}, nil
}

// NewInsecureK8sClient creates an insecure k8s client which is suitable
// to connect kubernetes api behind proxy
func NewInsecureK8sClient(apiURL string) K8sClient {
	return &k8sClient{
		host:       apiURL,
		httpClient: http.DefaultClient,
	}
}

func getEndpoints(client K8sClient, namespace, targetName string) (Endpoints, error) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/namespaces/%s/endpoints/%s",
		client.Host(), namespace, targetName))
	if err != nil {
		return Endpoints{}, err
	}
	req, err := client.GetRequest(u.String())
	if err != nil {
		return Endpoints{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return Endpoints{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Endpoints{}, fmt.Errorf("invalid response code %d", resp.StatusCode)
	}
	result := Endpoints{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result, err
}

func watchEndpoints(client K8sClient, namespace, targetName string) (watchInterface, error) {
	u, err := url.Parse(fmt.Sprintf("%s/api/v1/watch/namespaces/%s/endpoints/%s",
		client.Host(), namespace, targetName))
	if err != nil {
		return nil, err
	}
	req, err := client.GetRequest(u.String())
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, fmt.Errorf("invalid response code %d", resp.StatusCode)
	}
	return newStreamWatcher(resp.Body), nil
}

func getCurrentNamespaceOrDefault() string {
	ns, err := ioutil.ReadFile(kubernetesNamespaceFile)
	if err != nil {
		return defaultNamespace
	}
	return string(ns)
}
