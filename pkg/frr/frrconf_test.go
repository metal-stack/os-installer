package frr

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_ipPrefixListV2_String(t *testing.T) {
	tests := []struct {
		name string
		ipp  ipPrefixListV2
		want string
	}{
		{
			name: "simple ipv4",
			ipp: ipPrefixListV2{
				name:         "vrf10-import-from-vrf12",
				sequence:     100,
				accessPolicy: permit,
				prefix:       netip.MustParsePrefix("100.127.0.0/16"),
				len:          new(uint(32)),
			},
			want: "ip prefix-list vrf10-import-from-vrf12 seq 100 permit 100.127.0.0/16 le 32",
		},
		{
			name: "simple ipv6",
			ipp: ipPrefixListV2{
				name:         "vrf10-import-from-vrf12",
				sequence:     100,
				accessPolicy: permit,
				prefix:       netip.MustParsePrefix("2002::/64"),
				len:          new(uint(128)),
			},
			want: "ipv6 prefix-list vrf10-import-from-vrf12 seq 100 permit 2002::/64 le 128",
		},
		{
			name: "simple ipv6, seq empty",
			ipp: ipPrefixListV2{
				name:         "vrf10-import-from-vrf12",
				accessPolicy: permit,
				prefix:       netip.MustParsePrefix("2002::/64"),
				len:          new(uint(128)),
			},
			want: "ipv6 prefix-list vrf10-import-from-vrf12 permit 2002::/64 le 128",
		},
		{
			name: "simple ipv6, seq empty and len empty",
			ipp: ipPrefixListV2{
				name:         "vrf10-import-from-vrf12",
				accessPolicy: permit,
				prefix:       netip.MustParsePrefix("2002::/64"),
			},
			want: "ipv6 prefix-list vrf10-import-from-vrf12 permit 2002::/64",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ipp.String()

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_routeMapV2_String(t *testing.T) {
	tests := []struct {
		name string
		rm   routeMapV2
		want string
	}{
		{
			name: "simple deny only",
			rm: routeMapV2{
				name:        "vrf3982-import-map",
				matchPolicy: deny,
				order:       30,
			},
			want: `route-map vrf3982-import-map deny 30`,
		},
		{
			name: "simple v4",
			rm: routeMapV2{
				name:        "vrf3982-import-map",
				matchPolicy: permit,
				order:       10,
				matchers: []routeMapMatch{
					{matchType: matchTypeSourceVRF, sourceVrf: new(uint64(3981))},
					{matchType: matchTypeIPAddress, addressFamily: addressFamilyV4, prefixListName: new("vrf3982-import-from-vrf3981")},
				},
			},
			want: `route-map vrf3982-import-map permit 10
 match source-vrf vrf3981
 match ip address prefix-list vrf3982-import-from-vrf3981`,
		},
		{
			name: "simple v4 with no export",
			rm: routeMapV2{
				name:        "vrf3982-import-map",
				matchPolicy: permit,
				order:       10,
				matchers: []routeMapMatch{
					{matchType: matchTypeSourceVRF, sourceVrf: new(uint64(3981))},
					{matchType: matchTypeIPAddress, addressFamily: addressFamilyV4, prefixListName: new("vrf3982-import-from-vrf3981")},
					{matchType: matchTypeNoExport},
				},
			},
			want: `route-map vrf3982-import-map permit 10
 match source-vrf vrf3981
 match ip address prefix-list vrf3982-import-from-vrf3981
 set community additive no-export`,
		},
		{
			name: "simple v6 with no export",
			rm: routeMapV2{
				name:        "vrf3982-import-map",
				matchPolicy: permit,
				order:       10,
				matchers: []routeMapMatch{
					{matchType: matchTypeSourceVRF, sourceVrf: new(uint64(3982))},
					{matchType: matchTypeIPAddress, addressFamily: addressFamilyV6, prefixListName: new("vrf3982-import-from-vrf3981")},
					{matchType: matchTypeNoExport},
				},
			},
			want: `route-map vrf3982-import-map permit 10
 match source-vrf vrf3982
 match ipv6 address prefix-list vrf3982-import-from-vrf3981
 set community additive no-export`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.rm.String()
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_vrfV2_Router(t *testing.T) {
	tests := []struct {
		name       string
		asn        string
		routerId   string
		frrVersion frrVersion
		vrfV2      vrfV2
		want       string
	}{
		{
			name:       "simple",
			asn:        "123",
			routerId:   "123",
			frrVersion: frrVersion{Major: 10},
			vrfV2: vrfV2{
				vni: 345,
				vrf: 3981,
				importedVrfs: []importVrf{
					{
						vrf: new(uint64(3892)),
					},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.vrfV2.Router(tt.asn, tt.routerId, tt.frrVersion)
			require.Equal(t, tt.want, got)
		})
	}
}
