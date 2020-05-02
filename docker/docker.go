package corednsdocker

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/miekg/dns"
)

const pluginName = "docker"

var log = clog.NewWithPlugin(pluginName)

type docker struct {
	client   *client.Client
	networks []network
	next     plugin.Handler
	origins  []string
}

type network struct {
	name string
}

func (d docker) ServeDNS(ctx context.Context, w dns.ResponseWriter, msg *dns.Msg) (int, error) {
	state := request.Request{Req: msg, W: w}
	qname := state.Name()
	zone := plugin.Zones(d.origins).Matches(qname)
	if zone == "" {
		return d.next.ServeDNS(ctx, w, msg)
	}

	containers, err := d.client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return dns.RcodeServerFailure, fmt.Errorf(
			"listing docker containers: %w", err,
		)
	}

	var answers []dns.RR
findRecord:
	for _, container := range containers {
		for _, containerName := range container.Names {
			containerFQDN := fmt.Sprintf("%s.%s",
				strings.TrimPrefix(containerName, "/"),
				zone,
			)
			if containerFQDN == qname {
				for networkName, n := range container.NetworkSettings.Networks {
					for _, network := range d.networks {
						if network.name == networkName {
							answers = append(answers, &dns.AAAA{
								Hdr: dns.RR_Header{
									Name:   plugin.Name(zone).Normalize(),
									Rrtype: dns.TypeAAAA,
									Class:  dns.ClassINET,
									Ttl:    3600,
								},
								AAAA: net.ParseIP(n.GlobalIPv6Address),
							})
							break findRecord
						}
					}
				}
			}
		}
	}

	if len(answers) == 0 {
		resp := &dns.Msg{}
		resp.SetRcode(msg, dns.RcodeNameError)
		w.WriteMsg(resp)
		return dns.RcodeNameError, nil
	}

	resp := &dns.Msg{}
	resp.SetReply(msg)
	resp.Authoritative = true
	resp.Answer = answers
	w.WriteMsg(resp)
	return dns.RcodeSuccess, nil
}

func (d docker) Name() string { return pluginName }
