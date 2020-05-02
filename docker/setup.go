package corednsdocker

import (
	"github.com/caddyserver/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"github.com/docker/docker/client"
)

func init() { plugin.Register(pluginName, setup) }

func setup(c *caddy.Controller) error {
	docker, err := parseDocker(c)
	if err != nil {
		return err
	}

	docker.client, err = client.NewEnvClient()
	if err != nil {
		return c.Errf("creating docker client: %s", err)
	}

	for _, origin := range c.ServerBlockKeys {
		docker.origins = append(docker.origins, plugin.Host(origin).Normalize())
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		docker.next = next
		return docker
	})

	return nil
}

func parseDocker(c *caddy.Controller) (docker, error) {
	d := docker{}
	for c.Next() {
		args := c.RemainingArgs()
		if len(args) < 1 {
			return d, c.Err("no network provided")
		}
		if len(args) > 2 {
			return d, c.Err("too many arguments")
		}
		d.networks = append(d.networks, network{
			name: args[0],
		})
	}
	return d, nil
}
