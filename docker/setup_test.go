package corednsdocker

import (
	"fmt"
	"testing"

	"github.com/caddyserver/caddy"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		name             string
		corefile         string
		expectedNetworks []network
		shouldFail       bool
	}{
		{
			name:     "single network",
			corefile: `docker netname domain`,
			expectedNetworks: []network{
				{
					name: "netname",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := caddy.NewTestController("docker", test.corefile)
			d, err := parseDocker(c)
			if test.shouldFail {
				if err == nil {
					t.Fatalf("expected error, but got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no errors, but got: %v", err)
			}
			expected := fmt.Sprint(test.expectedNetworks)
			actual := fmt.Sprint(d.networks)
			if expected != actual {
				t.Fatalf("wrong networks: expected %s, got %s", expected, actual)
			}
		})
	}
}
