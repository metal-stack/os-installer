package nftables

import (
	"embed"
	"log/slog"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/os-installer/pkg/network"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	//go:embed test
	expectedNftableFiles embed.FS

	firewallAllocation = &apiv2.MachineAllocation{
		AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
		Networks: []*apiv2.MachineNetwork{
			{
				Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
				Prefixes:    []string{"10.0.16.0/22"},
				Ips:         []string{"10.0.16.2"},
				Vrf:         3981,
			},
			{
				Network:     "partition-storage",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
				Prefixes:    []string{"10.0.18.0/22"},
				Ips:         []string{"10.0.18.2"},
				Vrf:         3982,
				// FIXME clarify if this is required
				// NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:             "internet",
				NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Ips:                 []string{"185.1.2.3"},
				DestinationPrefixes: []string{"0.0.0.0/0"},
				Vrf:                 104009,
				NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:     "underlay",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
				Ips:         []string{"10.1.0.1"},
			},
			{
				Network:     "mpls",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Prefixes:    []string{"100.127.129.0/22"},
				Ips:         []string{"100.127.129.1"},
				Vrf:         104010,
				NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:     "internet-v6",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Ips:         []string{"2001::4"},
			},
		},
	}

	firewallWithVPNAllocation = &apiv2.MachineAllocation{
		AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
		Networks: []*apiv2.MachineNetwork{
			{
				Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
				Prefixes:    []string{"10.0.16.0/22"},
				Ips:         []string{"10.0.16.2"},
				Vrf:         3981,
			},
			{
				Network:     "partition-storage",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
				Prefixes:    []string{"10.0.18.0/22"},
				Ips:         []string{"10.0.18.2"},
				Vrf:         3982,
				// NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:             "internet",
				NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Ips:                 []string{"185.1.2.3"},
				DestinationPrefixes: []string{"0.0.0.0/0"},
				Vrf:                 104009,
				NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:     "underlay",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
				Ips:         []string{"10.1.0.1"},
			},
			{
				Network:     "mpls",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Prefixes:    []string{"100.127.129.0/22"},
				Ips:         []string{"100.127.129.1"},
				Vrf:         104010,
				NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:     "internet-v6",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Ips:         []string{"2001::4"},
			},
		},
		Vpn: &apiv2.MachineVPN{
			ControlPlaneAddress: "https://test.test.dev",
			AuthKey:             "abracadabra",
		},
	}
	firewallWithRulesAllocation = &apiv2.MachineAllocation{
		AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
		Networks: []*apiv2.MachineNetwork{
			{
				Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
				Prefixes:    []string{"10.0.16.0/22"},
				Ips:         []string{"10.0.16.2"},
				Vrf:         3981,
			},
			{
				Network:     "partition-storage",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
				Prefixes:    []string{"10.0.18.0/22"},
				Ips:         []string{"10.0.18.2"},
				Vrf:         3982,
				// NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:             "internet",
				NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Ips:                 []string{"185.1.2.3"},
				DestinationPrefixes: []string{"0.0.0.0/0"},
				Vrf:                 104009,
				NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:     "underlay",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
				Ips:         []string{"10.1.0.1"},
			},
			{
				Network:     "mpls",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Prefixes:    []string{"100.127.129.0/22"},
				Ips:         []string{"100.127.129.1"},
				Vrf:         104010,
				NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:     "internet-v6",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Ips:         []string{"2001::4"},
			},
		},
		FirewallRules: &apiv2.FirewallRules{
			Egress: []*apiv2.FirewallEgressRule{
				{
					Comment:  "allow apt update",
					Protocol: apiv2.IPProtocol_IP_PROTOCOL_TCP,
					Ports:    []uint32{443},
					To:       []string{"0.0.0.0/0", "1.2.3.4/32"},
				},
				{
					Comment:  "allow apt update v6",
					Protocol: apiv2.IPProtocol_IP_PROTOCOL_TCP,
					Ports:    []uint32{443},
					To:       []string{"::/0"},
				},
			},
			Ingress: []*apiv2.FirewallIngressRule{
				{
					Comment:  "allow incoming ssh",
					Protocol: apiv2.IPProtocol_IP_PROTOCOL_TCP,
					Ports:    []uint32{22},
					From:     []string{"2.3.4.0/24", "192.168.1.0/16"},
					To:       []string{"100.1.2.3/32", "100.1.2.4/32"},
				},
				{
					Comment:  "allow incoming ssh ipv6",
					Protocol: apiv2.IPProtocol_IP_PROTOCOL_TCP,
					Ports:    []uint32{22},
					From:     []string{"2001:db8::1/128"},
					To:       []string{"2001:db8:0:113::/64"},
				},
				{
					Protocol: apiv2.IPProtocol_IP_PROTOCOL_TCP,
					Ports:    []uint32{80, 443, 8080},
					From:     []string{"1.2.3.0/24", "192.168.0.0/16"},
				},
			},
		},
	}

	firewallSharedAllocation = &apiv2.MachineAllocation{
		AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
		Project:        "dd429d45-db03-4627-887f-bf7761d376a5",
		Networks: []*apiv2.MachineNetwork{
			{
				Network:     "partition-storage",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
				Prefixes:    []string{"10.0.18.0/22"},
				Ips:         []string{"10.0.18.2"},
				Vrf:         3982,
				Project:     new("dd429d45-db03-4627-887f-bf7761d376a5"),
				NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:             "internet",
				NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Ips:                 []string{"185.1.2.3"},
				DestinationPrefixes: []string{"0.0.0.0/0"},
				Vrf:                 104009,
				NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:     "underlay",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
				Ips:         []string{"10.1.0.1"},
			},
			{
				Network:     "internet-v6",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Ips:         []string{"2001::4"},
			},
		},
	}

	firewallIPv6Allocation = &apiv2.MachineAllocation{
		AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
		Networks: []*apiv2.MachineNetwork{
			{
				Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
				Prefixes:    []string{"2002::/64"},
				Ips:         []string{"2002::1"},
				Vrf:         3981,
			},
			{
				Network:     "partition-storage",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
				Prefixes:    []string{"10.0.18.0/22"},
				Ips:         []string{"10.0.18.2"},
				Vrf:         3982,
				// FIXME clarify if this is required
				// NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:             "internet",
				NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Prefixes:            []string{"2a02:c00:20::/45"},
				Ips:                 []string{"2a02:c00:20::1"},
				DestinationPrefixes: []string{"::/0"},
				Vrf:                 104009,
				NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:     "underlay",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
				Ips:         []string{"10.1.0.1"},
			},
			{
				Network:     "mpls",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Prefixes:    []string{"100.127.129.0/22"},
				Ips:         []string{"100.127.129.1"},
				Vrf:         104010,
				NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:     "internet-v6",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Ips:         []string{"2001::4"},
			},
		},
	}
)

func TestRender(t *testing.T) {
	tests := []struct {
		name           string
		allocation     *apiv2.MachineAllocation
		enableDNSProxy bool
		forwardPolicy  ForwardPolicy
		wantFilePath   string
		wantErr        error
	}{
		{
			name:           "render firewall, forward drop",
			allocation:     firewallAllocation,
			wantFilePath:   "nftrules",
			enableDNSProxy: false,
			forwardPolicy:  ForwardPolicyDrop,
			wantErr:        nil,
		},
		{
			name:           "render firewall, forward accept",
			allocation:     firewallAllocation,
			wantFilePath:   "nftrules_accept_forwarding",
			enableDNSProxy: false,
			forwardPolicy:  ForwardPolicyAccept,
			wantErr:        nil,
		},
		{
			name:           "render firewall with vpn",
			allocation:     firewallWithVPNAllocation,
			wantFilePath:   "nftrules_vpn",
			enableDNSProxy: false,
			forwardPolicy:  ForwardPolicyDrop,
			wantErr:        nil,
		},
		{
			name:           "render firewall with rules",
			allocation:     firewallWithRulesAllocation,
			wantFilePath:   "nftrules_with_rules",
			enableDNSProxy: false,
			forwardPolicy:  ForwardPolicyDrop,
			wantErr:        nil,
		},
		{
			name:           "render firewall shared",
			allocation:     firewallSharedAllocation,
			wantFilePath:   "nftrules_shared",
			enableDNSProxy: true,
			forwardPolicy:  ForwardPolicyDrop,
			wantErr:        nil,
		},
		{
			name:           "render firewall ipv6",
			allocation:     firewallIPv6Allocation,
			wantFilePath:   "nftrules_ipv6",
			enableDNSProxy: true,
			forwardPolicy:  ForwardPolicyDrop,
			wantErr:        nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}

			_, gotErr := Render(t.Context(), &Config{
				Log:            slog.Default(),
				fs:             fs,
				Network:        network.New(tt.allocation),
				EnableDNSProxy: tt.enableDNSProxy,
				ForwardPolicy:  tt.forwardPolicy,
				Validate:       false,
			})

			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile(nftrulesPath)
			require.NoError(t, err)

			assert.Equal(t, mustReadExpected(tt.wantFilePath), string(content))
		})
	}
}

func mustReadExpected(name string) string {
	tpl, err := expectedNftableFiles.ReadFile(path.Join("test", name))
	if err != nil {
		panic(err)
	}

	return string(tpl)
}
