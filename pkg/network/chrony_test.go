package network

import (
	"testing"

	"github.com/metal-stack/metal-go/api/models"
	mn "github.com/metal-stack/metal-lib/pkg/net"
	apiv1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChronyServiceEnabler_Enable(t *testing.T) {
	vrf := int64(104009)
	external := mn.External
	network := &models.V1MachineNetwork{Networktype: &external, Destinationprefixes: []string{IPv4ZeroCIDR}, Vrf: &vrf}
	tests := []struct {
		kb              config
		vrf             string
		isErrorExpected bool
	}{
		{
			kb:              config{InstallerConfig: apiv1.InstallerConfig{Networks: []*models.V1MachineNetwork{network}}},
			vrf:             "vrf104009",
			isErrorExpected: false,
		},
		{
			kb:              config{InstallerConfig: apiv1.InstallerConfig{Networks: []*models.V1MachineNetwork{}}},
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
