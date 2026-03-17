package services

import (
	"context"
	"errors"
	"log/slog"

	"github.com/metal-stack/os-installer/pkg/network"
	"github.com/metal-stack/os-installer/pkg/services/chrony"
	"github.com/metal-stack/os-installer/pkg/services/droptailer"
	firewallcontroller "github.com/metal-stack/os-installer/pkg/services/firewall-controller"
	nftablesexporter "github.com/metal-stack/os-installer/pkg/services/nftables-exporter"
	nodeexporter "github.com/metal-stack/os-installer/pkg/services/node-exporter"
	"github.com/metal-stack/os-installer/pkg/services/suricata"
	"github.com/metal-stack/os-installer/pkg/services/tailscale"
)

func WriteSystemdServices(ctx context.Context, log *slog.Logger, network *network.Network, machineUUID string) error {
	if network.IsMachine() {
		return nil
	}

	var (
		errs            []error
		defaultRouteVRF string
		tenantVRF       string
	)

	defaultRouteVRF, err := network.GetDefaultRouteNetworkVrfName()
	if err != nil {
		errs = append(errs, err)
	}
	tenantVRF, err = network.GetTenantNetworkVrfName()
	if err != nil {
		errs = append(errs, err)
	}

	// Droptailer
	if _, err = droptailer.WriteSystemdUnit(ctx, &droptailer.Config{
		Log:    log,
		Enable: true,
		Reload: false,
	}, &droptailer.TemplateData{
		Comment:   "created from os-installer",
		TenantVrf: tenantVRF,
	}); err != nil {
		errs = append(errs, err)
	}

	// Chrony
	if _, err = chrony.WriteSystemdUnit(ctx, &chrony.Config{
		Log:              log,
		Enable:           false,
		Reload:           false,
		ChronyConfigPath: "",
	}, &chrony.TemplateData{
		NTPServers: network.NTPServers(),
	}, defaultRouteVRF); err != nil {
		errs = append(errs, err)
	}

	// firewall-controller
	if _, err = firewallcontroller.WriteSystemdUnit(ctx, &firewallcontroller.Config{
		Log:    log,
		Enable: true,
		Reload: false,
	}, &firewallcontroller.TemplateData{
		Comment:         "created from os-installer",
		DefaultRouteVrf: defaultRouteVRF,
	}); err != nil {
		errs = append(errs, err)
	}

	// nftables-exporter
	if _, err := nftablesexporter.WriteSystemdUnit(ctx, &nftablesexporter.Config{
		Log:    log,
		Enable: true,
		Reload: false,
	}, &nftablesexporter.TemplateData{
		Comment: "created from os-installer",
	}); err != nil {
		errs = append(errs, err)
	}

	// node-exporter
	if _, err := nodeexporter.WriteSystemdUnit(ctx, &nodeexporter.Config{
		Log:    log,
		Enable: true,
		Reload: false,
	}, &nodeexporter.TemplateData{
		Comment: "created from os-installer",
	}); err != nil {
		errs = append(errs, err)
	}

	// suricata
	if _, err := suricata.WriteSystemdUnit(ctx, &suricata.Config{
		Log:    log,
		Enable: true,
		Reload: false,
	}, &suricata.TemplateData{
		Interface:       "TODO",
		DefaultRouteVrf: defaultRouteVRF,
	}); err != nil {
		errs = append(errs, err)
	}

	// tailscale
	if network.HasVpn() {
		vpn := network.Vpn()
		if _, err := tailscale.WriteSystemdUnit(ctx, &tailscale.Config{
			Log:    log,
			Enable: true,
			Reload: false,
		}, &tailscale.TemplateData{
			Comment:         "created from os-installer",
			DefaultRouteVrf: defaultRouteVRF,
			MachineID:       machineUUID,
			AuthKey:         vpn.AuthKey,
			Address:         vpn.ControlPlaneAddress,
		}); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
