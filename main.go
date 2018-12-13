package main

import (
	"encoding/json"
	"github.com/google/tcpproxy"
	"io/ioutil"
	"log"
	"strings"
	"sync"
)

var (
	empty = struct{}{}
	pm    = &proxyManager{
		httpProxyRule:  make(map[string]struct{}),
		httpsProxyRule: make(map[string]struct{}),
		stopCh:         make(chan struct{}),
		config:         new(Config),
	}
)

type ProxyProto string

var (
	ProtoHTTP  ProxyProto = "http"
	ProtoHTTPS ProxyProto = "https"
)

type ProxyInstance struct {
	Listen string     `json:"listen"`
	Proto  ProxyProto `json:"proto"`
	Hosts  []string   `json:"hosts"`
}

type Config struct {
	Instances []ProxyInstance `json:"instances"`
}

type proxyManager struct {
	config         *Config
	httpProxyRule  map[string]struct{}
	httpsProxyRule map[string]struct{}
	p              tcpproxy.Proxy
	mtx            sync.Mutex
	stopCh         chan struct{}
}

func (pm *proxyManager) Stop() {
	pm.stopCh <- empty
}

//func (pm *proxyManager) AddHTTPRule(host string) {
//	pm.mtx.Lock()
//
//	if _, ok := pm.httpProxyRule[host]; !ok {
//		pm.httpProxyRule[host] = empty
//		pm.reload()
//	}
//	pm.mtx.Unlock()
//}
//
//func (pm *proxyManager) AddHTTPSRule(host string) {
//	pm.mtx.Lock()
//
//	if _, ok := pm.httpsProxyRule[host]; !ok {
//		pm.httpsProxyRule[host] = empty
//		pm.reload()
//	}
//	pm.mtx.Unlock()
//}

func (pm *proxyManager) reload() {
	pm.stop()
	pm.start()
}

func (pm *proxyManager) stop() {
	pm.p.Close()
}

func proxyHost(host string, port string) string {
	if len(strings.Split(host, ":")) == 1 {
		return host + ":" + port
	}
	return host
}

func (pm *proxyManager) start() {
	pm.p = tcpproxy.Proxy{}

	for _, instance := range pm.config.Instances {
		if instance.Proto == ProtoHTTP {
			for _, host := range instance.Hosts {
				pm.p.AddHTTPHostRoute(instance.Listen, host, tcpproxy.To(proxyHost(host, "80")))
				log.Printf(`p.AddHTTPHostRoute(":80", %v, tcpproxy.To(%v))`, host, proxyHost(host, "80"))
			}
		}

		if instance.Proto == ProtoHTTPS {
			for _, host := range instance.Hosts {
				pm.p.AddSNIRoute(instance.Listen, host, tcpproxy.To(proxyHost(host, "443")))
				log.Printf(`p.AddHTTPSHostRoute(":443", %v, tcpproxy.To(%v))`, host, proxyHost(host, "443"))
			}
		}

	}

	//for host, _ := range pm.httpProxyRule {
	//	log.Printf(`p.AddHTTPHostRoute(":80", %v, tcpproxy.To(%v))`, host, proxyHost(host, "80"))
	//	pm.p.AddHTTPHostRoute(":80", host, tcpproxy.To(proxyHost(host, "80")))
	//}

	//for host, _ := range pm.httpsProxyRule {
	//	log.Printf(`p.AddHTTPHostRoute(":443", %v, tcpproxy.To(%v))`, host, proxyHost(host, "443"))
	//	//pm.p.AddHTTPHostRoute(":443", host, tcpproxy.To(proxyHost(host, "443")))
	//	pm.p.AddSNIRoute(":443", host, tcpproxy.To(proxyHost(host, "443")))
	//}
	//pm.p.AddRoute(":80", tcpproxy.To("127.0.0.1:8888"))
	//pm.p.AddRoute(":443", tcpproxy.To("127.0.0.1:8883"))
	//log.Println("proxy server http route:", pm.httpProxyRule)
	//log.Println("proxy server https route:", pm.httpsProxyRule)
	pm.p.Start()
	log.Println("proxy server start ...")

}

func (pm *proxyManager) Run() {

	data, err := ioutil.ReadFile("/etc/dagent/dagent.json")
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, pm.config); err != nil {
		panic(err)
	}
	log.Println("DEUBG:", pm.config)

	pm.start()

	<-pm.stopCh
}

//func dealHTTP(rw http.ResponseWriter, r *http.Request) {
//	pm.AddHTTPRule(r.Host)
//	rw.WriteHeader(http.StatusServiceUnavailable)
//
//}

//func dealHTTPS(rw http.ResponseWriter, r *http.Request) {
//	pm.AddHTTPSRule(r.Host)
//	rw.WriteHeader(http.StatusServiceUnavailable)
//
//}

func main() {

	//go http.ListenAndServe("127.0.0.1:8888", http.HandlerFunc(dealHTTP))
	//go http.ListenAndServe("127.0.0.1:8883", http.HandlerFunc(dealHTTPS))

	pm.Run()

}
