package network

import (
	"testing"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChronyServiceEnabler_Enable(t *testing.T) {
	vrf := uint64(104009)
	external := apiv2.NetworkType_NETWORK_TYPE_EXTERNAL
	network := &apiv2.MachineNetwork{NetworkType: external, DestinationPrefixes: []string{IPv4ZeroCIDR}, Vrf: vrf}
	tests := []struct {
		kb              config
		vrf             string
		isErrorExpected bool
	}{
		{
			kb:              config{MachineAllocation: &apiv2.MachineAllocation{Networks: []*apiv2.MachineNetwork{network}}},
			vrf:             "vrf104009",
			isErrorExpected: false,
		},
		{
			kb:              config{MachineAllocation: &apiv2.MachineAllocation{Networks: []*apiv2.MachineNetwork{}}},
			vrf:             "",
			isErrorExpected: true,
		},
	}

	for _, tt := range tests {
		e, err := newChronyServiceEnabler(tt.kb)
		if tt.isErrorExpected {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
		assert.Equal(t, tt.vrf, e.vrf)
	}
}
