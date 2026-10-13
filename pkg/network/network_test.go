package network_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/os-installer/pkg/network"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/stretchr/testify/require"
)

func TestNetwork_MTU(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       int
	}{
		{
			name:       "firewall",
			allocation: &apiv2.MachineAllocation{AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL},
			want:       9216,
		},
		{
			name:       "machine",
			allocation: &apiv2.MachineAllocation{AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE},
			want:       9000,
		},
		{
			name:       "unknown",
			allocation: &apiv2.MachineAllocation{AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_UNSPECIFIED},
			want:       9000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got := n.MTU()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_Hostname(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       string
	}{
		{
			name:       "with hostname",
			allocation: &apiv2.MachineAllocation{Hostname: "metal"},
			want:       "metal",
		},
		{
			name:       "without hostname",
			allocation: &apiv2.MachineAllocation{Hostname: ""},
			want:       "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got := n.Hostname()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_IsMachine(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       bool
	}{
		{
			name:       "firewall",
			allocation: &apiv2.MachineAllocation{AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL},
			want:       false,
		},
		{
			name:       "machine",
			allocation: &apiv2.MachineAllocation{AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE},
			want:       true,
		},
		{
			name:       "unknown",
			allocation: &apiv2.MachineAllocation{AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_UNSPECIFIED},
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got := n.IsMachine()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_HasVpn(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       bool
	}{
		{
			name:       "firewall with vpn",
			allocation: &apiv2.MachineAllocation{Vpn: &apiv2.MachineVPN{AuthKey: "secret"}},
			want:       true,
		},
		{
			name:       "firewall with vpn but not authkey",
			allocation: &apiv2.MachineAllocation{Vpn: &apiv2.MachineVPN{}},
			want:       false,
		},
		{
			name:       "firewall without vpn",
			allocation: &apiv2.MachineAllocation{},
			want:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got := n.HasVpn()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_NTPServers(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       []string
	}{
		{
			name:       "with one ntp",
			allocation: &apiv2.MachineAllocation{NtpServers: []*apiv2.NTPServer{{Address: "ntp.pool.org"}}},
			want:       []string{"ntp.pool.org"},
		},
		{
			name:       "with two ntp",
			allocation: &apiv2.MachineAllocation{NtpServers: []*apiv2.NTPServer{{Address: "ntp.pool.org"}, {Address: "ntp2.pool.org"}}},
			want:       []string{"ntp.pool.org", "ntp2.pool.org"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got := n.NTPServers()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_LoopbackCIDRs(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       []string
		wantErr    error
	}{
		{
			name: "firewall",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want:    []string{"10.1.0.1/32"},
			wantErr: nil,
		},
		{
			name: "machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want:    []string{"10.0.16.2/32", "10.0.18.2/32"},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got, err := n.LoopbackCIDRs()
			if diff := cmp.Diff(tt.wantErr, err, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_UnderlayNetwork(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       *apiv2.MachineNetwork
		wantErr    error
	}{
		{
			name: "firewall",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want: &apiv2.MachineNetwork{
				Network:     "underlay",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
				Asn:         4200003073,
				Ips:         []string{"10.1.0.1"},
				Prefixes:    []string{"10.0.12.0/22"},
			},
			wantErr: nil,
		},
		{
			name: "machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("no underlay network present in network allocation"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got, err := n.UnderlayNetwork()
			if diff := cmp.Diff(tt.wantErr, err, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_PrivatePrimaryNetwork(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       *apiv2.MachineNetwork
		wantErr    error
	}{
		{
			name: "firewall",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want: &apiv2.MachineNetwork{
				Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
				Project:     new("project-a"),
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
				Prefixes:    []string{"10.0.16.0/22"},
				Ips:         []string{"10.0.16.2"},
				Vrf:         3981,
				Asn:         4200003073,
			},
			wantErr: nil,
		},
		{
			name: "machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want: &apiv2.MachineNetwork{
				Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
				Project:     new("project-a"),
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
				Prefixes:    []string{"10.0.16.0/22"},
				Ips:         []string{"10.0.16.2"},
				Vrf:         3981,
				Asn:         4200003073,
			},
			wantErr: nil,
		},
		{
			name: "storage machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				Project:        "project-b",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want: &apiv2.MachineNetwork{
				Network:     "partition-storage",
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
				Project:     new("project-b"),
				Prefixes:    []string{"10.0.18.0/22"},
				Ips:         []string{"10.0.18.2"},
				Vrf:         3982,
				Asn:         4200003073,
			},
			wantErr: nil,
		},
		{
			name: "storage machine in wrong project",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				Project:        "project-b",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-c"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("no private primary network present in network allocation"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got, err := n.PrivatePrimaryNetwork()
			if diff := cmp.Diff(tt.wantErr, err, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_PrivateSecondarySharedNetworks(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       []*apiv2.MachineNetwork
	}{
		{
			name: "firewall",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want: []*apiv2.MachineNetwork{
				{
					Network:     "partition-storage",
					NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
					Project:     new("project-b"),
					Prefixes:    []string{"10.0.18.0/22"},
					Ips:         []string{"10.0.18.2"},
					Vrf:         3982,
					Asn:         4200003073,
				},
			},
		},
		{
			name: "machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want: []*apiv2.MachineNetwork{
				{
					Network:     "partition-storage",
					NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
					Project:     new("project-b"),
					Prefixes:    []string{"10.0.18.0/22"},
					Ips:         []string{"10.0.18.2"},
					Vrf:         3982,
					Asn:         4200003073,
				},
			},
		},
		{
			name: "storage machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				Project:        "project-b",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got := n.PrivateSecondarySharedNetworks()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_PrivatePrimaryIPs(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       []string
		wantErr    error
	}{
		{
			name: "firewall",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want:    []string{"10.1.0.1"},
			wantErr: nil,
		},
		{
			name: "machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want:    []string{"10.0.16.2"},
			wantErr: nil,
		},
		{
			name: "storage machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				Project:        "project-b",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want:    []string{"10.0.18.2"},
			wantErr: nil,
		},
		{
			name: "storage machine in wrong project",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				Project:        "project-b",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-c"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("no private primary ip present in network allocation"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got, err := n.PrivatePrimaryIPs()
			if diff := cmp.Diff(tt.wantErr, err, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_PrivatePrimaryNetworksPrefixes(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       []string
		wantErr    error
	}{
		{
			name: "firewall",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want:    []string{"10.0.16.0/22"},
			wantErr: nil,
		},
		{
			name: "machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want:    []string{"10.0.16.0/22"},
			wantErr: nil,
		},
		{
			name: "storage machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				Project:        "project-b",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want:    []string{"10.0.16.0/22"},
			wantErr: nil,
		},
		{
			name: "storage machine in wrong project",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				Project:        "project-b",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-c"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("no private primary networks present in network allocation"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got, err := n.PrivatePrimaryNetworksPrefixes()
			if diff := cmp.Diff(tt.wantErr, err, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_VxlanIDs(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       []uint64
	}{
		{
			name: "firewall",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want: []uint64{3981, 3982, 104009, 104010},
		},
		{
			name: "machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want: []uint64{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got := n.VxlanIDs()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_EVPNIfaces(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       []network.EvpnIface
		wantErr    error
	}{
		{
			name: "firewall",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3", "185.1.2.4"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want: []network.EvpnIface{
				{
					Network: "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
					CIDRs:   []string{"10.0.16.2/32"},
					VlanID:  1000,
					VrfID:   3981,
				},
				{
					Network: "partition-storage",
					CIDRs:   []string{"10.0.18.2/32"},
					VlanID:  1001,
					VrfID:   3982,
				},
				{
					Network: "internet",
					CIDRs:   []string{"185.1.2.3/32", "185.1.2.4/32"},
					VlanID:  1002,
					VrfID:   104009,
				},
				{
					Network: "mpls",
					CIDRs:   []string{"100.127.129.1/32"},
					VlanID:  1004,
					VrfID:   104010,
				},
			},
			wantErr: nil,
		},
		{
			name: "machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("no evpn interfaces supported on machines"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got, err := n.EVPNIfaces()
			if diff := cmp.Diff(tt.wantErr, err, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_GetNetworks(t *testing.T) {
	tests := []struct {
		name        string
		allocation  *apiv2.MachineAllocation
		networkType apiv2.NetworkType
		want        []*apiv2.MachineNetwork
	}{
		{
			name: "firewall external networks",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3", "185.1.2.4"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			networkType: apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
			want: []*apiv2.MachineNetwork{
				{
					Network:             "internet",
					NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
					Ips:                 []string{"185.1.2.3", "185.1.2.4"},
					Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
					DestinationPrefixes: []string{"0.0.0.0/0"},
					Vrf:                 104009,
					Asn:                 4200003073,
					NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
				},
				{
					Network:             "mpls",
					NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
					Prefixes:            []string{"100.127.129.0/24"},
					Ips:                 []string{"100.127.129.1"},
					DestinationPrefixes: []string{"100.127.1.0/24"},
					Vrf:                 104010,
					Asn:                 4200003073,
					NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
				},
			},
		},

		{
			name: "firewall underlay networks",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3", "185.1.2.4"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			networkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
			want: []*apiv2.MachineNetwork{
				{
					Network:     "underlay",
					NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
					Asn:         4200003073,
					Ips:         []string{"10.1.0.1"},
					Prefixes:    []string{"10.0.12.0/22"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got := n.GetNetworks(tt.networkType)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_GetExternalNetworkVrfNames(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       []string
	}{
		{
			name: "firewall external networks",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3", "185.1.2.4"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want: []string{"vrf104009", "vrf104010"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got := n.GetExternalNetworkVrfNames()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_GetDefaultRouteNetwork(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       *apiv2.MachineNetwork
		wantErr    error
	}{
		{
			name: "firewall",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3", "185.1.2.4"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want: &apiv2.MachineNetwork{
				Network:             "internet",
				NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Ips:                 []string{"185.1.2.3", "185.1.2.4"},
				Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
				DestinationPrefixes: []string{"0.0.0.0/0"},
				Vrf:                 104009,
				Asn:                 4200003073,
				NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			wantErr: nil,
		},
		{
			name: "machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want:    nil,
			wantErr: errors.New("no network which provides a default route found"),
		},

		{
			name: "firewall dualstack",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"2a02:c00:20::1", "185.1.2.3"},
						Prefixes:            []string{"185.1.2.0/24", "2a02:c00:20::/45"},
						DestinationPrefixes: []string{"::/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want: &apiv2.MachineNetwork{
				Network:             "internet",
				NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Ips:                 []string{"2a02:c00:20::1", "185.1.2.3"},
				Prefixes:            []string{"185.1.2.0/24", "2a02:c00:20::/45"},
				DestinationPrefixes: []string{"::/0"},
				Vrf:                 104009,
				Asn:                 4200003073,
				NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got, err := n.GetDefaultRouteNetwork()
			if diff := cmp.Diff(tt.wantErr, err, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_GetDefaultRouteNetworkVrfName(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       string
		wantErr    error
	}{
		{
			name: "firewall",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3", "185.1.2.4"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want:    "vrf104009",
			wantErr: nil,
		},
		{
			name: "machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want:    "",
			wantErr: errors.New("no network which provides a default route found"),
		},

		{
			name: "firewall dualstack",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"2a02:c00:20::1", "185.1.2.3"},
						Prefixes:            []string{"185.1.2.0/24", "2a02:c00:20::/45"},
						DestinationPrefixes: []string{"::/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want:    "vrf104009",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got, err := n.GetDefaultRouteNetworkVrfName()
			if diff := cmp.Diff(tt.wantErr, err, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNetwork_GetTenantNetworkVrfName(t *testing.T) {
	tests := []struct {
		name       string
		allocation *apiv2.MachineAllocation
		want       string
		wantErr    error
	}{
		{
			name: "firewall",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "firewall",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
				Project:        "project-a",
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
					{
						Network:             "internet",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Ips:                 []string{"185.1.2.3", "185.1.2.4"},
						Prefixes:            []string{"185.1.2.0/24", "185.27.0.0/22"},
						DestinationPrefixes: []string{"0.0.0.0/0"},
						Vrf:                 104009,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
					{
						Network:     "underlay",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_UNDERLAY,
						Asn:         4200003073,
						Ips:         []string{"10.1.0.1"},
						Prefixes:    []string{"10.0.12.0/22"},
					},
					{
						Network:             "mpls",
						NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
						Prefixes:            []string{"100.127.129.0/24"},
						Ips:                 []string{"100.127.129.1"},
						DestinationPrefixes: []string{"100.127.1.0/24"},
						Vrf:                 104010,
						Asn:                 4200003073,
						NatType:             apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
					},
				},
			},
			want:    "vrf3981",
			wantErr: nil,
		},
		{
			name: "machine",
			allocation: &apiv2.MachineAllocation{
				Hostname:       "machine",
				AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
				Networks: []*apiv2.MachineNetwork{
					{
						Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
						Project:     new("project-a"),
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
						Prefixes:    []string{"10.0.16.0/22"},
						Ips:         []string{"10.0.16.2"},
						Vrf:         3981,
						Asn:         4200003073,
					},
					{
						Network:     "partition-storage",
						NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED,
						Project:     new("project-b"),
						Prefixes:    []string{"10.0.18.0/22"},
						Ips:         []string{"10.0.18.2"},
						Vrf:         3982,
						Asn:         4200003073,
					},
				},
			},
			want: "vrf3981",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := network.New(tt.allocation)
			got, err := n.GetTenantNetworkVrfName()
			if diff := cmp.Diff(tt.wantErr, err, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}
			require.Equal(t, tt.want, got)
		})
	}
}
