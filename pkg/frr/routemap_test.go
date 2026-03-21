package frr

import (
	"fmt"
	"log/slog"
	"net/netip"
	"testing"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/os-installer/pkg/network"
	"github.com/stretchr/testify/require"
)

type (
	testnetwork struct {
		vrf          string
		prefixes     []importPrefix
		destinations []importPrefix
	}
	importSettings struct {
		ImportPrefixes         []importPrefix
		ImportPrefixesNoExport []importPrefix
	}
)

var (
	defaultRoute           = importPrefix{Prefix: netip.MustParsePrefix("0.0.0.0/0"), Policy: permit, SourceVRF: inetVrf}
	defaultRoute6          = importPrefix{Prefix: netip.MustParsePrefix("::/0"), Policy: permit, SourceVRF: inetVrf}
	externalVrf            = "vrf104010"
	externalNet            = importPrefix{Prefix: netip.MustParsePrefix("100.127.129.0/24"), Policy: permit, SourceVRF: externalVrf}
	externalDestinationNet = importPrefix{Prefix: netip.MustParsePrefix("100.127.1.0/24"), Policy: permit, SourceVRF: externalVrf}
	privateVrf             = "vrf3981"
	privateNet             = importPrefix{Prefix: netip.MustParsePrefix("10.0.16.0/22"), Policy: permit, SourceVRF: privateVrf}
	privateNet6            = importPrefix{Prefix: netip.MustParsePrefix("2002::/64"), Policy: permit, SourceVRF: privateVrf}
	sharedVrf              = "vrf3982"
	sharedNet              = importPrefix{Prefix: netip.MustParsePrefix("10.0.18.0/22"), Policy: permit, SourceVRF: sharedVrf}
	inetVrf                = "vrf104009"
	inetNet1               = importPrefix{Prefix: netip.MustParsePrefix("185.1.2.0/24"), Policy: permit, SourceVRF: inetVrf}
	inetNet2               = importPrefix{Prefix: netip.MustParsePrefix("185.27.0.0/22"), Policy: permit, SourceVRF: inetVrf}
	inetNet6               = importPrefix{Prefix: netip.MustParsePrefix("2a02:c00:20::/45"), Policy: permit, SourceVRF: inetVrf}
	publicDefaultNet       = importPrefix{Prefix: netip.MustParsePrefix("185.1.2.3/32"), Policy: deny, SourceVRF: inetVrf}
	publicDefaultNetIPv6   = importPrefix{Prefix: netip.MustParsePrefix("2a02:c00:20::1/128"), Policy: deny, SourceVRF: inetVrf}

	private = testnetwork{
		vrf:      privateVrf,
		prefixes: []importPrefix{privateNet},
	}

	private6 = testnetwork{
		vrf:      privateVrf,
		prefixes: []importPrefix{privateNet6},
	}

	inet = testnetwork{
		vrf:          inetVrf,
		prefixes:     []importPrefix{inetNet1, inetNet2},
		destinations: []importPrefix{defaultRoute},
	}

	inet6 = testnetwork{
		vrf:          inetVrf,
		prefixes:     []importPrefix{inetNet6},
		destinations: []importPrefix{defaultRoute6},
	}
	dualstack = testnetwork{
		vrf:          inetVrf,
		prefixes:     []importPrefix{inetNet1, inetNet6},
		destinations: []importPrefix{defaultRoute6},
	}
	external = testnetwork{
		vrf:          externalVrf,
		destinations: []importPrefix{externalDestinationNet},
		prefixes:     []importPrefix{externalNet},
	}

	shared = testnetwork{
		vrf:      sharedVrf,
		prefixes: []importPrefix{sharedNet},
	}
)

func Test_importRulesForNetwork(t *testing.T) {
	// FIXME enable once we understand whats actually tested
	t.Skip()
	tests := []struct {
		name  string
		input *apiv2.MachineAllocation
		want  map[string]map[string]importSettings
	}{
		{
			name:  "standard firewall with private primary unshared network, private secondary shared network, internet and mpls",
			input: firewallAllocation,
			want: map[string]map[string]importSettings{
				// The target VRF
				private.vrf: {
					// Imported VRFs with their restrictions
					inet.vrf: importSettings{
						ImportPrefixes: concatPfxSlices(inet.destinations, []importPrefix{publicDefaultNet}, inet.prefixes),
					},
					external.vrf: importSettings{
						ImportPrefixes: concatPfxSlices(external.destinations, external.prefixes),
					},
					shared.vrf: importSettings{
						ImportPrefixes: shared.prefixes,
					},
				},
				shared.vrf: {
					private.vrf: importSettings{
						ImportPrefixes: concatPfxSlices(private.prefixes, leakFrom(shared.prefixes, private.vrf)),
					},
				},
				inet.vrf: {
					private.vrf: importSettings{
						ImportPrefixes:         leakFrom(inet.prefixes, private.vrf),
						ImportPrefixesNoExport: private.prefixes,
					},
				},
				external.vrf: {
					private.vrf: importSettings{
						ImportPrefixes:         leakFrom(external.prefixes, private.vrf),
						ImportPrefixesNoExport: private.prefixes,
					},
				},
			},
		},
		{
			name:  "firewall of a shared private network (shared/storage firewall)",
			input: firewallSharedAllocation,
			want: map[string]map[string]importSettings{
				shared.vrf: {
					inet.vrf: importSettings{
						ImportPrefixes: concatPfxSlices(inet.destinations, []importPrefix{publicDefaultNet}, inet.prefixes),
					},
				},
				inet.vrf: {
					shared.vrf: importSettings{
						ImportPrefixes:         leakFrom(inet.prefixes, shared.vrf),
						ImportPrefixesNoExport: shared.prefixes,
					},
				},
			},
		},
		{
			name:  "firewall with ipv6 private network and ipv6 internet network",
			input: firewallIPv6Allocation,
			want: map[string]map[string]importSettings{
				private6.vrf: {
					inet6.vrf: importSettings{
						ImportPrefixes: concatPfxSlices(inet6.destinations, []importPrefix{publicDefaultNetIPv6}, inet6.prefixes),
					},
					external.vrf: importSettings{
						ImportPrefixes: concatPfxSlices(external.destinations, external.prefixes),
					},
					shared.vrf: importSettings{
						ImportPrefixes: shared.prefixes,
					},
				},
				shared.vrf: {
					private6.vrf: importSettings{
						ImportPrefixes: concatPfxSlices(private6.prefixes, leakFrom(shared.prefixes, private6.vrf)),
					},
				},
				inet6.vrf: {
					private6.vrf: importSettings{
						ImportPrefixes:         leakFrom(inet6.prefixes, private6.vrf),
						ImportPrefixesNoExport: private6.prefixes,
					},
				},
				external.vrf: {
					private6.vrf: importSettings{
						ImportPrefixes:         leakFrom(external.prefixes, private6.vrf),
						ImportPrefixesNoExport: private6.prefixes,
					},
				},
			},
		},
		{
			name:  "firewall with ipv6 private network and dualstack internet network",
			input: firewallAllocationDualStack,
			want: map[string]map[string]importSettings{
				private6.vrf: {
					inet6.vrf: importSettings{
						ImportPrefixes: concatPfxSlices(inet6.destinations, []importPrefix{publicDefaultNetIPv6, publicDefaultNet}, dualstack.prefixes),
					},
					external.vrf: importSettings{
						ImportPrefixes: concatPfxSlices(external.destinations, external.prefixes),
					},
					shared.vrf: importSettings{
						ImportPrefixes: shared.prefixes,
					},
				},
				shared.vrf: {
					private6.vrf: importSettings{
						ImportPrefixes: concatPfxSlices(private6.prefixes, leakFrom(shared.prefixes, private6.vrf)),
					},
				},
				inet6.vrf: {
					private6.vrf: importSettings{
						ImportPrefixes:         leakFrom(dualstack.prefixes, private6.vrf),
						ImportPrefixesNoExport: private6.prefixes,
					},
				},
				external.vrf: {
					private6.vrf: importSettings{
						ImportPrefixes:         leakFrom(external.prefixes, private6.vrf),
						ImportPrefixesNoExport: private6.prefixes,
					},
				},
			},
		},
	}
	log := slog.Default()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{
				Log:     log,
				Network: network.New(tt.input),
			}
			for _, network := range tt.input.Networks {
				got, err := importRulesForNetwork(cfg, network)
				require.NoError(t, err)
				if got == nil {
					continue
				}
				gotBySourceVrf := got.bySourceVrf()
				targetVrf := fmt.Sprintf("vrf%d", network.Vrf)
				want := tt.want[targetVrf]

				require.Equal(t, want, gotBySourceVrf)
			}
		})
	}
}

func (i *importRule) bySourceVrf() map[string]importSettings {
	r := map[string]importSettings{}
	for _, vrf := range i.ImportVRFs {
		r[vrf] = importSettings{}
	}

	for _, pfx := range i.ImportPrefixes {
		e := r[pfx.SourceVRF]
		e.ImportPrefixes = append(e.ImportPrefixes, pfx)
		r[pfx.SourceVRF] = e
	}

	for _, pfx := range i.ImportPrefixesNoExport {
		e := r[pfx.SourceVRF]
		e.ImportPrefixesNoExport = append(e.ImportPrefixesNoExport, pfx)
		r[pfx.SourceVRF] = e
	}

	return r
}

func leakFrom(pfxs []importPrefix, sourceVrf string) []importPrefix {
	r := []importPrefix{}
	for _, e := range pfxs {
		i := e
		i.SourceVRF = sourceVrf
		r = append(r, i)
	}
	return r
}
