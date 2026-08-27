package control

import (
	"context"
	"fmt"

	dnsdao "github.com/gofurry/gofurry-nav-collector/collector/dns/dao"
	dnsservice "github.com/gofurry/gofurry-nav-collector/collector/dns/service"
	"github.com/gofurry/gofurry-nav-collector/collector/execution"
	httpdao "github.com/gofurry/gofurry-nav-collector/collector/http/dao"
	httpservice "github.com/gofurry/gofurry-nav-collector/collector/http/service"
	lightdao "github.com/gofurry/gofurry-nav-collector/collector/lightprobe/dao"
	lightservice "github.com/gofurry/gofurry-nav-collector/collector/lightprobe/service"
	"github.com/gofurry/gofurry-nav-collector/collector/observation"
	pingdao "github.com/gofurry/gofurry-nav-collector/collector/ping/dao"
	pingservice "github.com/gofurry/gofurry-nav-collector/collector/ping/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProtocolExecutor struct {
	ping  *pingservice.Runner
	http  *httpservice.Runner
	dns   *dnsservice.Runner
	light *lightservice.Runner
}

func NewProtocolExecutor(pool *pgxpool.Pool) *ProtocolExecutor {
	observations := observation.NewDAO(pool)
	return &ProtocolExecutor{
		ping:  pingservice.InitPingOnStart(pingdao.New(pool), observations),
		http:  httpservice.InitHTTPOnStart(httpdao.New(pool), observations),
		dns:   dnsservice.InitDNSOnStart(dnsdao.New(pool), observations),
		light: lightservice.InitLightProbeOnStart(lightdao.New(pool), observations),
	}
}

func (e *ProtocolExecutor) Run(ctx context.Context, protocol string, request execution.Request) error {
	request.Context = ctx
	var run func()
	switch protocol {
	case observation.ProtocolPing:
		run = e.ping.Ping
	case observation.ProtocolHTTP:
		run = e.http.Request
	case observation.ProtocolDNS:
		run = e.dns.ParseDNS
	case observation.ProtocolRDAP:
		run = e.light.RunRDAP
	case observation.ProtocolRobots:
		run = e.light.RunRobots
	case observation.ProtocolSecurityTXT:
		run = e.light.RunSecurityTXT
	case observation.ProtocolLLMSTXT:
		run = e.light.RunLLMSTXT
	case observation.ProtocolPageAssets:
		run = e.light.RunPageAssets
	case observation.ProtocolPortCheck:
		run = e.light.RunPortCheck
	case observation.ProtocolWAFCanary:
		run = e.light.RunWAFCanary
	default:
		return fmt.Errorf("unsupported Nav protocol %q", protocol)
	}
	return execution.With(protocol, request, run)
}

func (e *ProtocolExecutor) RetainLegacyLogs() {
	go e.ping.Delete()
	go e.http.Delete()
	go e.dns.Delete()
}

func (e *ProtocolExecutor) Close() { dnsservice.CloseGeoDB() }
