package oscommon

import (
	"context"
	"fmt"

	"github.com/metal-stack/os-installer/pkg/frr"
	"github.com/metal-stack/os-installer/pkg/interfaces"
	"github.com/metal-stack/os-installer/pkg/nftables"
)

func (d *DefaultOS) ConfigureNetwork(ctx context.Context) error {
	if err := interfaces.ConfigureInterfaces(ctx, &interfaces.Config{
		Log:     d.log,
		Network: d.network,
		Nics:    d.details.Nics,
	}); err != nil {
		return fmt.Errorf("error configuring interfaces: %w", err)
	}

	if _, err := frr.Render(ctx, &frr.Config{
		Log:      d.log,
		Reload:   false,
		Validate: true,
		Network:  d.network,
	}); err != nil {
		return fmt.Errorf("unable to render frr config: %w", err)
	}

	if d.network.IsMachine() {
		return nil
	}

	if _, err := nftables.Render(ctx, &nftables.Config{
		Log:            d.log,
		Reload:         false,
		Network:        d.network,
		EnableDNSProxy: false,
		ForwardPolicy:  nftables.ForwardPolicyDrop,
	}); err != nil {
		return fmt.Errorf("unable to render nftables config: %w", err)
	}

	return nil
}
