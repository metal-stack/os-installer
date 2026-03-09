package network

import (
	"fmt"
	"net/netip"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/samber/lo"
)

const (
	// mtuFirewall defines the value for MTU specific to the needs of a firewall. VXLAN requires higher MTU.
	mtuFirewall = 9216
	// mtuMachine defines the value for MTU specific to the needs of a machine.
	mtuMachine = 9000
)

type Network struct {
	allocation *apiv2.MachineAllocation
}

func New(allocation *apiv2.MachineAllocation) *Network {
	return &Network{
		allocation: allocation,
	}
}

func (n *Network) MTU() int {
	if n.allocation.AllocationType == apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL {
		return mtuFirewall
	}

	return mtuMachine
}

func (n *Network) LoopbackCIDRs() (cidrs []string, err error) {
	var ips []string

	if n.allocation.AllocationType == apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL {
		ips, err = loFirewallIps(n.allocation.Networks)
		if err != nil {
			return nil, err
		}
	} else {
		ips, err = loMachineIps(n.allocation.Networks)
		if err != nil {
			return nil, err
		}
	}

	for _, ip := range ips {
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			return nil, err
		}

		cidrs = append(cidrs, fmt.Sprintf("%s/%d", addr.String(), addr.BitLen()))
	}

	return
}

func (n *Network) PrivatePrimaryIPs() ([]string, error) {
	if n.allocation.AllocationType == apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL {
		for _, nw := range n.allocation.Networks {
			if nw.NetworkType == apiv2.NetworkType_NETWORK_TYPE_UNDERLAY {
				return nw.Ips, nil
			}
		}

		return nil, fmt.Errorf("no private primary ip present in network allocation")
	}

	for _, nw := range n.allocation.Networks {
		if nw.NetworkType == apiv2.NetworkType_NETWORK_TYPE_CHILD {
			return nw.Ips, nil
		}
	}

	for _, nw := range n.allocation.Networks {
		if nw.Project == nil {
			continue
		}

		if nw.NetworkType == apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED && *nw.Project == n.allocation.Project {
			return nw.Ips, nil
		}
	}

	return nil, fmt.Errorf("no private primary ip present in network allocation")
}

func (n *Network) VxlanIDs() (ids []uint64) {
	if n.allocation.AllocationType == apiv2.MachineAllocationType_MACHINE_ALLOCATION_TYPE_FIREWALL {
		for _, nw := range n.allocation.Networks {
			if nw.Vrf > 0 {
				ids = append(ids, nw.Vrf)
			}
		}
	}

	ids = lo.Uniq(ids)

	return
}

func loFirewallIps(networks []*apiv2.MachineNetwork) (ips []string, err error) {
	for _, nw := range networks {
		switch nw.NetworkType {
		case apiv2.NetworkType_NETWORK_TYPE_UNDERLAY:
			ips = append(ips, nw.Ips...)
		}
	}

	return
}

func loMachineIps(networks []*apiv2.MachineNetwork) (ips []string, err error) {
	for _, nw := range networks {
		switch nw.NetworkType {
		case apiv2.NetworkType_NETWORK_TYPE_CHILD, apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED:
			ips = append(ips, nw.Ips...)
		case apiv2.NetworkType_NETWORK_TYPE_EXTERNAL:
			ips = append(ips, nw.Ips...)
		}
	}

	return
}
