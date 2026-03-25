package frr

import (
	"log/slog"
	"net/netip"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/os-installer/pkg/network"
	"github.com/stretchr/testify/require"
)

func Test_importRulesForNetwork(t *testing.T) {
	log := slog.Default()

	tests := []struct {
		name    string
		cfg     *Config
		network *apiv2.MachineNetwork
		want    *importRule
	}{
		{
			name: "primary private network of a firewall",
			cfg: &Config{
				Log:     log,
				Network: network.New(firewallAllocation),
			},
			network: firewallAllocation.Networks[0],
			want: &importRule{
				TargetVRF: vrfNameOf(firewallAllocation.Networks[0]),
				ImportVRFs: []string{
					vrfNameOf(firewallAllocation.Networks[2]),
					vrfNameOf(firewallAllocation.Networks[4]),
					vrfNameOf(firewallAllocation.Networks[1]),
				},
				ImportPrefixes: []importPrefix{
					{Prefix: netip.MustParsePrefix("0.0.0.0/0"), Policy: permit, SourceVRF: vrfNameOf(firewallAllocation.Networks[2])},
					{Prefix: netip.MustParsePrefix("100.127.1.0/24"), Policy: permit, SourceVRF: vrfNameOf(firewallAllocation.Networks[4])},
					{Prefix: netip.MustParsePrefix("185.1.2.3/32"), Policy: deny, SourceVRF: vrfNameOf(firewallAllocation.Networks[2])},
					{Prefix: netip.MustParsePrefix("185.1.2.0/24"), Policy: permit, SourceVRF: vrfNameOf(firewallAllocation.Networks[2])},
					{Prefix: netip.MustParsePrefix("185.27.0.0/22"), Policy: permit, SourceVRF: vrfNameOf(firewallAllocation.Networks[2])},
					{Prefix: netip.MustParsePrefix("100.127.129.0/24"), Policy: permit, SourceVRF: vrfNameOf(firewallAllocation.Networks[4])},
					{Prefix: netip.MustParsePrefix("10.0.18.0/22"), Policy: permit, SourceVRF: vrfNameOf(firewallAllocation.Networks[1])},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := importRulesForNetwork(tt.cfg, tt.network)
			require.NoError(t, err)

			if diff := cmp.Diff(tt.want, got, cmp.AllowUnexported(netip.Prefix{}), cmpopts.IgnoreUnexported(netip.Addr{})); diff != "" {
				t.Errorf("importRulesForNetwork() diff = %s", diff)
			}
		})
	}
}
