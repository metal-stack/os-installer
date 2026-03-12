package frr

import (
	"embed"
	"log/slog"
	"path"
	"testing"

	"github.com/Masterminds/semver/v3"
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
	expectedFrrFiles embed.FS

	firewallAllocation = &apiv2.MachineAllocation{
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
				// FIXME clarify if this is required
				// NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
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
	}

	firewallAllocationDualStack = &apiv2.MachineAllocation{
		Hostname:       "firewall",
		AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
		Project:        "project-a",
		Networks: []*apiv2.MachineNetwork{
			{
				Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
				Project:     new("project-a"),
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
				Prefixes:    []string{"2002::/64"},
				Ips:         []string{"2002::1"},
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
				// FIXME clarify if this is required
				// NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
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
	}

	firewallFrr9Allocation = &apiv2.MachineAllocation{
		Hostname:       "firewall",
		AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
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
				// FIXME clarify if this is required
				// NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
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
	}

	firewallFrr10Allocation = &apiv2.MachineAllocation{
		Hostname:       "firewall",
		AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
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
				// FIXME clarify if this is required
				// NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
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
	}

	firewallSharedAllocation = &apiv2.MachineAllocation{
		Hostname:       "firewall",
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
		},
	}

	firewallIPv6Allocation = &apiv2.MachineAllocation{
		Hostname:       "firewall",
		AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL,
		Networks: []*apiv2.MachineNetwork{
			{
				Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
				Project:     new("project-a"),
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
				Prefixes:    []string{"2002::/64"},
				Ips:         []string{"2002::1"},
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
				// FIXME clarify if this is required
				// NatType:     apiv2.NATType_NAT_TYPE_IPV4_MASQUERADE,
			},
			{
				Network:             "internet",
				NetworkType:         apiv2.NetworkType_NETWORK_TYPE_EXTERNAL,
				Ips:                 []string{"2a02:c00:20::1"},
				Prefixes:            []string{"2a02:c00:20::/45"},
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
	}

	machineAllocation = &apiv2.MachineAllocation{
		Hostname:       "machine",
		Project:        "project-a",
		AllocationType: apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_MACHINE,
		Networks: []*apiv2.MachineNetwork{
			{
				Network:     "379d294d-22e8-4aed-82e1-62c6c2f08d6a",
				Project:     new("project-a"),
				NetworkType: apiv2.NetworkType_NETWORK_TYPE_CHILD,
				Prefixes:    []string{"10.0.16.0/22"},
				Ips:         []string{"10.0.17.2"},
				Vrf:         3981,
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
	}
)

func TestRender(t *testing.T) {
	tests := []struct {
		name         string
		allocation   *apiv2.MachineAllocation
		frrVersion   *semver.Version
		wantFilePath string
		wantErr      error
	}{
		{
			name:         "render firewall",
			allocation:   firewallAllocation,
			wantFilePath: "frr.conf.firewall",
			wantErr:      nil,
		},
		{
			name:         "render firewall, dualstack",
			allocation:   firewallAllocationDualStack,
			wantFilePath: "frr.conf.firewall_dualstack",
			wantErr:      nil,
		},
		{
			name:         "render firewall frr-9",
			allocation:   firewallFrr9Allocation,
			frrVersion:   semver.MustParse("9.0.1"),
			wantFilePath: "frr.conf.firewall_frr-9",
			wantErr:      nil,
		},
		{
			name:         "render firewall frr-10",
			allocation:   firewallFrr10Allocation,
			frrVersion:   semver.MustParse("10.4.1"),
			wantFilePath: "frr.conf.firewall_frr-10",
			wantErr:      nil,
		},
		{
			name:         "render firewall shared",
			allocation:   firewallSharedAllocation,
			wantFilePath: "frr.conf.firewall_shared",
			wantErr:      nil,
		},
		{
			name:         "render firewall ipv6",
			allocation:   firewallIPv6Allocation,
			wantFilePath: "frr.conf.firewall_ipv6",
			wantErr:      nil,
		},
		{
			name:         "render machine",
			allocation:   machineAllocation,
			wantFilePath: "frr.conf.machine",
			wantErr:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}

			_, gotErr := Render(t.Context(), &Config{
				Log:        slog.Default(),
				fs:         fs,
				Network:    network.New(tt.allocation),
				FRRVersion: tt.frrVersion,
				Validate:   false,
			})

			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			content, err := fs.ReadFile(frrConfigPath)
			require.NoError(t, err)

			assert.Equal(t, mustReadExpected(tt.wantFilePath), string(content))
		})
	}
}

func mustReadExpected(name string) string {
	tpl, err := expectedFrrFiles.ReadFile(path.Join("test", name))
	if err != nil {
		panic(err)
	}

	return string(tpl)
}
