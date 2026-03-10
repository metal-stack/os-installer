package frr

import (
	"fmt"
	"net/netip"
	"sort"
	"strings"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	mn "github.com/metal-stack/metal-lib/pkg/net"
	nwutil "github.com/metal-stack/os-installer/pkg/network"
)

const (
	// Permit defines an access policy that allows access.
	Permit AccessPolicy = "permit"
	// Deny defines an access policy that forbids access.
	Deny AccessPolicy = "deny"
)

// AccessPolicy is a type that represents a policy to manage access roles.
type (
	AccessPolicy string

	importPrefix struct {
		Prefix    netip.Prefix
		Policy    AccessPolicy
		SourceVRF string
	}

	importRule struct {
		TargetVRF              string
		ImportVRFs             []string
		ImportPrefixes         []importPrefix
		ImportPrefixesNoExport []importPrefix
	}

	ImportSettings struct {
		ImportPrefixes         []importPrefix
		ImportPrefixesNoExport []importPrefix
	}
)

func (i *importRule) bySourceVrf() map[string]ImportSettings {
	r := map[string]ImportSettings{}
	for _, vrf := range i.ImportVRFs {
		r[vrf] = ImportSettings{}
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

func importRulesForNetwork(cfg *Config, network *apiv2.MachineNetwork) (*importRule, error) {
	if network.NetworkType == apiv2.NetworkType_NETWORK_TYPE_UNDERLAY {
		return nil, nil
	}

	vrfName := vrfNameOf(network)
	i := importRule{
		TargetVRF: vrfName,
	}
	privatePrimaryNet, err := cfg.Network.PrivatePrimaryNetwork()
	if err != nil {
		return nil, err
	}

	externalNets := cfg.Network.GetNetworks(apiv2.NetworkType_NETWORK_TYPE_EXTERNAL)
	privateSecondarySharedNets := cfg.Network.GetNetworks(mn.PrivateSecondaryShared)

	switch network.NetworkType {
	case mn.PrivatePrimaryUnshared:
		fallthrough
	// case mn.PrivatePrimaryShared:
	case apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED:
		// reach out from private network into public networks
		i.ImportVRFs = vrfNamesOf(externalNets)
		i.ImportPrefixes = getDestinationPrefixes(externalNets)

		// deny public address of default network
		defaultNet, err := cfg.Network.GetDefaultRouteNetwork()
		if err != nil {
			return nil, err
		}
		for _, ip := range defaultNet.Ips {
			if parsed, err := netip.ParseAddr(ip); err == nil {
				var bl = 32
				if parsed.Is6() {
					bl = 128
				}
				i.ImportPrefixes = append(i.ImportPrefixes, importPrefix{
					Prefix:    netip.PrefixFrom(parsed, bl),
					Policy:    Deny,
					SourceVRF: vrfNameOf(defaultNet),
				})
			}
		}

		// permit external routes
		i.ImportPrefixes = append(i.ImportPrefixes, prefixesOfNetworks(externalNets)...)

		// reach out from private network into shared private networks
		i.ImportVRFs = append(i.ImportVRFs, vrfNamesOf(privateSecondarySharedNets)...)
		i.ImportPrefixes = append(i.ImportPrefixes, prefixesOfNetworks(privateSecondarySharedNets)...)

		// reach out from private network to destination prefixes of private secondays shared networks
		for _, n := range privateSecondarySharedNets {
			for _, pfx := range n.DestinationPrefixes {
				ppfx := netip.MustParsePrefix(pfx)
				isThere := false
				for _, i := range i.ImportPrefixes {
					if i.Prefix == ppfx {
						isThere = true
					}
				}
				if !isThere {
					i.ImportPrefixes = append(i.ImportPrefixes, importPrefix{
						Prefix:    ppfx,
						Policy:    Permit,
						SourceVRF: vrfNameOf(n),
					})
				}
			}
		}
	case mn.PrivateSecondaryShared:
		// reach out from private shared networks into private primary network
		i.ImportVRFs = []string{vrfNameOf(privatePrimaryNet)}
		i.ImportPrefixes = concatPfxSlices(prefixesOfNetwork(privatePrimaryNet, vrfNameOf(privatePrimaryNet)), prefixesOfNetwork(network, vrfNameOf(privatePrimaryNet)))

		// import destination prefixes of dmz networks from external networks
		if len(network.DestinationPrefixes) > 0 {
			for _, pfx := range network.DestinationPrefixes {
				for _, e := range externalNets {
					importExternalNet := false
					for _, epfx := range e.DestinationPrefixes {
						if pfx == epfx {
							importExternalNet = true
							i.ImportPrefixes = append(i.ImportPrefixes, importPrefix{
								Prefix:    netip.MustParsePrefix(pfx),
								Policy:    Permit,
								SourceVRF: vrfNameOf(e),
							})
						}
					}
					if importExternalNet {
						i.ImportVRFs = append(i.ImportVRFs, vrfNameOf(e))
						i.ImportPrefixes = append(i.ImportPrefixes, prefixesOfNetwork(e, vrfNameOf(e))...)
					}
				}
			}
		}
	case apiv2.NetworkType_NETWORK_TYPE_EXTERNAL:
		// reach out from public into private and other public networks
		i.ImportVRFs = []string{vrfNameOf(privatePrimaryNet)}
		i.ImportPrefixes = prefixesOfNetwork(network, vrfNameOf(privatePrimaryNet))

		nets := []*apiv2.MachineNetwork{privatePrimaryNet}

		if nwutil.ContainsDefaultRoute(network.DestinationPrefixes) {
			for _, r := range privateSecondarySharedNets {
				if nwutil.ContainsDefaultRoute(r.DestinationPrefixes) {
					nets = append(nets, r)
					i.ImportVRFs = append(i.ImportVRFs, vrfNameOf(r))
				}
			}
		}
		i.ImportPrefixesNoExport = prefixesOfNetworks(nets)
	}

	return &i, nil
}

func (i *importRule) prefixLists() []IPPrefixList {
	var result []IPPrefixList
	seed := IPPrefixListSeqSeed
	afs := []apiv2.NetworkAddressFamily{apiv2.NetworkAddressFamily_NETWORK_ADDRESS_FAMILY_V4, apiv2.NetworkAddressFamily_NETWORK_ADDRESS_FAMILY_V6}
	for _, af := range afs {
		pfxList := prefixLists(i.ImportPrefixesNoExport, &af, false, seed, i.TargetVRF)
		result = append(result, pfxList...)

		seed = IPPrefixListSeqSeed + len(result)
		result = append(result, prefixLists(i.ImportPrefixes, &af, true, seed, i.TargetVRF)...)
	}

	return result
}

func prefixLists(
	prefixes []importPrefix,
	af *apiv2.NetworkAddressFamily,
	isExported bool,
	seed int,
	vrf string,
) []IPPrefixList {
	var result []IPPrefixList
	for _, p := range prefixes {
		if *af == apiv2.NetworkAddressFamily_NETWORK_ADDRESS_FAMILY_V4 && !p.Prefix.Addr().Is4() {
			continue
		}

		if *af == apiv2.NetworkAddressFamily_NETWORK_ADDRESS_FAMILY_V6 && !p.Prefix.Addr().Is6() {
			continue
		}

		specs := p.buildSpecs(seed)
		for _, spec := range specs {
			// self-importing prefixes is nonsense
			if vrf == p.SourceVRF {
				continue
			}
			name := p.name(vrf, isExported)
			prefixList := IPPrefixList{
				Name:          name,
				Spec:          spec,
				AddressFamily: af,
				SourceVRF:     p.SourceVRF,
			}
			result = append(result, prefixList)
		}
		seed++
	}
	return result
}

func concatPfxSlices(pfxSlices ...[]importPrefix) []importPrefix {
	res := []importPrefix{}
	for _, pfxSlice := range pfxSlices {
		res = append(res, pfxSlice...)
	}
	return res
}

func stringSliceToIPPrefix(s []string, sourceVrf string) []importPrefix {
	var result []importPrefix
	for _, e := range s {
		ipp, err := netip.ParsePrefix(e)
		if err != nil {
			continue
		}
		result = append(result, importPrefix{
			Prefix:    ipp,
			Policy:    Permit,
			SourceVRF: sourceVrf,
		})
	}
	return result
}

func getDestinationPrefixes(networks []*apiv2.MachineNetwork) []importPrefix {
	var result []importPrefix
	for _, network := range networks {
		result = append(result, stringSliceToIPPrefix(network.DestinationPrefixes, vrfNameOf(network))...)
	}
	return result
}

func prefixesOfNetworks(networks []*apiv2.MachineNetwork) []importPrefix {
	var result []importPrefix
	for _, network := range networks {
		result = append(result, prefixesOfNetwork(network, vrfNameOf(network))...)
	}
	return result
}

func prefixesOfNetwork(network *apiv2.MachineNetwork, sourceVrf string) []importPrefix {
	return stringSliceToIPPrefix(network.Prefixes, sourceVrf)
}

func vrfNameOf(n *apiv2.MachineNetwork) string {
	return fmt.Sprintf("vrf%d", n.Vrf)
}

func vrfNamesOf(networks []*apiv2.MachineNetwork) []string {
	var result []string
	for _, n := range networks {
		result = append(result, vrfNameOf(n))
	}

	return result
}

func byName(prefixLists []IPPrefixList) map[string]IPPrefixList {
	byName := map[string]IPPrefixList{}
	for _, prefixList := range prefixLists {
		if _, isPresent := byName[prefixList.Name]; isPresent {
			continue
		}

		byName[prefixList.Name] = prefixList
	}

	return byName
}

func (i *importRule) routeMaps() []RouteMap {
	var result []RouteMap

	order := RouteMapOrderSeed
	byName := byName(i.prefixLists())

	names := []string{}
	for n := range byName {
		names = append(names, n)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(names)))

	for _, n := range names {
		prefixList := byName[n]

		matchVrf := fmt.Sprintf("match source-vrf %s", prefixList.SourceVRF)
		matchPfxList := fmt.Sprintf("match %s address prefix-list %s", prefixList.AddressFamily, n)
		entries := []string{matchVrf, matchPfxList}
		if strings.HasSuffix(n, IPPrefixListNoExportSuffix) {
			entries = append(entries, "set community additive no-export")
		}

		routeMap := RouteMap{
			Name:    routeMapName(i.TargetVRF),
			Policy:  string(Permit),
			Order:   order,
			Entries: entries,
		}
		order += RouteMapOrderSeed

		result = append(result, routeMap)
	}

	routeMap := RouteMap{
		Name:   routeMapName(i.TargetVRF),
		Policy: string(Deny),
		Order:  order,
	}

	result = append(result, routeMap)

	return result
}

func routeMapName(vrfName string) string {
	return vrfName + "-import-map"
}

func (i *importPrefix) buildSpecs(seq int) []string {
	var result []string
	var spec string

	if i.Prefix.Bits() == 0 {
		spec = fmt.Sprintf("%s %s", i.Policy, i.Prefix)

	} else {
		spec = fmt.Sprintf("seq %d %s %s le %d", seq, i.Policy, i.Prefix, i.Prefix.Addr().BitLen())
	}

	result = append(result, spec)

	return result
}

func (i *importPrefix) name(targetVrf string, isExported bool) string {
	suffix := ""

	if i.Prefix.Addr().Is6() {
		suffix = "-ipv6"
	}
	if !isExported {
		suffix += IPPrefixListNoExportSuffix
	}

	return fmt.Sprintf("%s-import-from-%s%s", targetVrf, i.SourceVRF, suffix)
}
