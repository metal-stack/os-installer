package network

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"

	apiv1 "github.com/metal-stack/os-installer/api/v1"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/metal-stack/v"

	"gopkg.in/yaml.v3"
)

const (
	// VLANOffset defines a number to start with when creating new VLAN IDs.
	VLANOffset = 1000
)

type (
	// config was generated with: https://mengzhuo.github.io/yaml-to-go/.
	// It represents the input yaml that is needed to render network configuration files.
	config struct {
		*apiv2.MachineAllocation
		machineDetails *apiv1.MachineDetails
		log            *slog.Logger
	}
)

// New creates a new instance of this type.
func New(log *slog.Logger, path string) (*config, error) {
	log.Info("loading", "path", path)

	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	allocation := &apiv2.MachineAllocation{}
	err = yaml.Unmarshal(f, &allocation)

	if err != nil {
		return nil, err
	}

	return &config{
		MachineAllocation: allocation,
		log:               log,
	}, nil
}

// Validate validates the containing information depending on the demands of the bare metal type.
func (c config) Validate(kind BareMetalType) error {
	if len(c.Networks) == 0 {
		return errors.New("expectation at least one network is present failed")
	}

	if !c.containsSinglePrivatePrimary() {
		return errors.New("expectation exactly one 'private: true' network is present failed")
	}

	if kind == Firewall {
		if !c.allNonUnderlayNetworksHaveNonZeroVRF() {
			return errors.New("networks with 'underlay: false' must contain a value of 'vrf' as it is used for BGP")
		}

		if !c.containsSingleUnderlay() {
			return errors.New("expectation exactly one underlay network is present failed")
		}

		if !c.containsAnyPublicNetwork() {
			return errors.New("expectation at least one public network (private: false, " +
				"underlay: false) is present failed")
		}

		for _, net := range c.GetNetworks(apiv2.NetworkType_NETWORK_TYPE_EXTERNAL) {
			if len(net.DestinationPrefixes) == 0 {
				return errors.New("non-private, non-underlay networks must contain destination prefix(es) to make " +
					"any sense of it")
			}
		}

		if c.isAnyNAT() && len(c.getPrivatePrimaryNetwork().Prefixes) == 0 {
			return errors.New("private network must not lack prefixes since nat is required")
		}
	}

	net := c.getPrivatePrimaryNetwork()

	if kind == Firewall {
		net = c.getUnderlayNetwork()
	}

	if len(net.Ips) == 0 {
		return errors.New("at least one IP must be present to be considered as LOOPBACK IP (" +
			"'private: true' network IP for machine, 'underlay: true' network IP for firewall")
	}

	if net.Asn <= 0 {
		return errors.New("'asn' of private (machine) resp. underlay (firewall) network must not be missing")
	}

	if len(c.machineDetails.Nics) == 0 {
		return errors.New("at least one 'nics/nic' definition must be present")
	}

	if !c.nicsContainValidMACs() {
		return errors.New("each 'nic' definition must contain a valid 'mac'")
	}

	return nil
}

func (c config) containsAnyPublicNetwork() bool {
	if len(c.GetNetworks(apiv2.NetworkType_NETWORK_TYPE_EXTERNAL)) > 0 {
		return true
	}
	return false
}

func (c config) containsSinglePrivatePrimary() bool {
	return c.containsSingleNetworkOf(apiv2.NetworkType_NETWORK_TYPE_CHILD) != c.containsSingleNetworkOf(apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED)
}

func (c config) containsSingleUnderlay() bool {
	return c.containsSingleNetworkOf(apiv2.NetworkType_NETWORK_TYPE_UNDERLAY)
}

func (c config) containsSingleNetworkOf(t apiv2.NetworkType) bool {
	possibleNetworks := c.GetNetworks(t)
	return len(possibleNetworks) == 1
}

// CollectIPs collects IPs of the given networks.
func (c config) CollectIPs(types ...apiv2.NetworkType) []string {
	var result []string

	networks := c.GetNetworks(types...)
	for _, network := range networks {
		result = append(result, network.Ips...)
	}

	return result
}

// GetNetworks returns all networks present.
func (c config) GetNetworks(types ...apiv2.NetworkType) []*apiv2.MachineNetwork {
	var result []*apiv2.MachineNetwork

	for _, t := range types {
		for _, n := range c.Networks {
			if n.NetworkType == apiv2.NetworkType_NETWORK_TYPE_UNSPECIFIED {
				continue
			}
			if n.NetworkType == t {
				result = append(result, n)
			}
		}
	}

	return result
}

func (c config) isAnyNAT() bool {
	for _, net := range c.Networks {
		if net.NatType != apiv2.NATType_NAT_TYPE_NONE || net.NatType != apiv2.NATType_NAT_TYPE_UNSPECIFIED {
			return true
		}
	}

	return false
}

func (c config) getPrivatePrimaryNetwork() *apiv2.MachineNetwork {
	return c.GetNetworks(apiv2.NetworkType_NETWORK_TYPE_CHILD, apiv2.NetworkType_NETWORK_TYPE_CHILD_SHARED)[0]
}

func (c config) getUnderlayNetwork() *apiv2.MachineNetwork {
	// Safe access since validation ensures there is exactly one.
	return c.GetNetworks(apiv2.NetworkType_NETWORK_TYPE_UNDERLAY)[0]
}

func (c config) GetDefaultRouteNetwork() *apiv2.MachineNetwork {
	externalNets := c.GetNetworks(apiv2.NetworkType_NETWORK_TYPE_EXTERNAL)
	for _, network := range externalNets {
		if containsDefaultRoute(network.DestinationPrefixes) {
			return network
		}
	}

	// Not supported anymore
	// privateSecondarySharedNets := c.GetNetworks(mn.PrivateSecondaryShared)
	// for _, network := range privateSecondarySharedNets {
	// 	if containsDefaultRoute(network.DestinationPrefixes) {
	// 		return network
	// 	}
	// }

	return nil
}

func (c config) getDefaultRouteVRFName() (string, error) {
	if network := c.GetDefaultRouteNetwork(); network != nil {
		return vrfNameOf(network), nil
	}

	return "", fmt.Errorf("there is no network providing a default (0.0.0.0/0) route")
}

func (c config) nicsContainValidMACs() bool {
	for _, nic := range c.machineDetails.Nics {
		if nic.Mac == "" {
			return false
		}

		if _, err := net.ParseMAC(nic.Mac); err != nil {
			c.log.Error("invalid mac", "mac", nic.Mac)
			return false
		}
	}

	return true
}

func (c config) allNonUnderlayNetworksHaveNonZeroVRF() bool {
	for _, net := range c.Networks {
		if net.NetworkType == apiv2.NetworkType_NETWORK_TYPE_UNDERLAY {
			continue
		}

		if net.Vrf <= 0 {
			return false
		}
	}

	return true
}

func versionHeader(uuid string) string {
	version := v.V.String()
	if os.Getenv("GO_ENV") == "testing" {
		version = ""
	}
	return fmt.Sprintf("# This file was auto generated for machine: '%s' by app version %s.\n# Do not edit.",
		uuid, version)
}
