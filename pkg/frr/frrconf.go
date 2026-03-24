package frr

import (
	"fmt"
	"net/netip"
	"strings"
)

const (
	matchTypeSourceVRF matchType = "source-vrf"
	matchTypeIPAddress matchType = "address"
	matchTypeNoExport  matchType = "set community additive no-export"

	addressFamilyV4 addressFamily = "ip"
	addressFamilyV6 addressFamily = "ipv6"
)

type (
	matchType     string
	addressFamily string

	ipPrefixListV2 struct {
		name         string
		sequence     int // is omitted of 0
		accessPolicy accessPolicy
		prefix       netip.Prefix
		len          *uint // only maxlen supported at the moment
	}

	routeMapV2 struct {
		name        string
		matchPolicy accessPolicy
		order       int // is omitted of 0
		matchers    []routeMapMatch
	}

	routeMapMatch struct {
		matchType      matchType
		sourceVrf      *uint64
		prefixListName *string
		addressFamily  addressFamily
	}

	importVrf struct {
		vrf      *uint64
		routeMap *routeMapV2
	}

	vrfV2 struct {
		vni           uint64
		vrf           uint64
		ipPrefixLists []ipPrefixListV2
		routeMaps     []routeMapV2
		importedVrfs  []importVrf
	}
)

func (ipp *ipPrefixListV2) String() string {
	var (
		af  string
		seq string
		len string
	)

	if ipp.prefix.Addr().Is4() {
		af = string(addressFamilyV4)
	}
	if ipp.prefix.Addr().Is6() {
		af = string(addressFamilyV6)
	}
	if ipp.sequence > 0 {
		seq = fmt.Sprintf(" seq %d", ipp.sequence)
	}

	if ipp.len != nil {
		len = fmt.Sprintf(" le %d", *ipp.len)
	}

	return fmt.Sprintf("%s prefix-list %s%s %s %s%s", af, ipp.name, seq, ipp.accessPolicy, ipp.prefix.String(), len)
}

func (rm *routeMapV2) String() string {
	var (
		result []string
	)

	result = append(result, fmt.Sprintf("route-map %s %s %d", rm.name, rm.matchPolicy, rm.order))

	for _, match := range rm.matchers {
		var matchLine string
		switch match.matchType {
		case matchTypeSourceVRF:
			matchLine = fmt.Sprintf(" match %s vrf%d", string(matchTypeSourceVRF), *match.sourceVrf)
		case matchTypeIPAddress:
			matchLine = fmt.Sprintf(" match %s %s prefix-list %s", match.addressFamily, string(matchTypeIPAddress), *match.prefixListName)
		case matchTypeNoExport:
			matchLine = fmt.Sprintf(" %s", string(matchTypeNoExport))
		}

		result = append(result, matchLine)
	}

	return strings.Join(result, "\n")
}

func (ivrf *importVrf) String() string {
	if ivrf.vrf != nil {
		return fmt.Sprintf(" import vrf vrf%d", *ivrf.vrf)
	}
	if ivrf.routeMap != nil {
		return fmt.Sprintf(" import vrf route-map %s", ivrf.routeMap.name)
	}
	return ""
}

func (v *vrfV2) L3VNI() string {
	return fmt.Sprintf(`vrf vrf%d
 vni %d
exit vrf
`, v.vrf, v.vni)
}

const routerTemplate = `
router bgp %s vrf vrf%d
 bgp router-id %s
 bgp bestpath as-path multipath-relax
 !
 address-family ipv4 unicast
  redistribute connected
 %s
 exit-address-family
 !
 address-family ipv6 unicast
  redistribute connected
 %s
 exit-address-family
 !
 address-family l2vpn evpn
  advertise ipv4 unicast
  advertise ipv6 unicast
 exit-address-family
 `

func (v *vrfV2) Router(asn, routerId string, frrVersion frrVersion) string {
	var importedVrfs []string

	for _, ivrf := range v.importedVrfs {
		importedVrfs = append(importedVrfs, ivrf.String())
	}

	ivrfs := strings.Join(importedVrfs, "\n")

	return fmt.Sprintf(routerTemplate, asn, v.vrf, routerId, ivrfs, ivrfs)
}
