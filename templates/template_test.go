package templates

import (
	_ "embed"
	"os"
	"testing"

	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	"github.com/stretchr/testify/require"
)

func TestDefaultChronyTemplate(t *testing.T) {
	defaultNTPServer := "time.cloudflare.com"
	ntpServers := []*apiv2.NTPServer{
		{
			Address: defaultNTPServer,
		},
	}

	rendered := renderToString(t, Chrony{NTPServers: ntpServers})
	expected := readExpected(t, "test_data/defaultntp/chrony.conf")

	require.Equal(t, expected, rendered, "Wanted: %s\nGot: %s", expected, rendered)
}

func TestCustomChronyTemplate(t *testing.T) {
	customNTPServer := "custom.1.ntp.org"
	ntpServers := []*apiv2.NTPServer{
		{
			Address: customNTPServer,
		},
	}

	rendered := renderToString(t, Chrony{NTPServers: ntpServers})
	expected := readExpected(t, "test_data/customntp/chrony.conf")

	require.Equal(t, expected, rendered, "Wanted: %s\nGot: %s", expected, rendered)
}

func readExpected(t *testing.T, e string) string {
	ex, err := os.ReadFile(e)
	require.NoError(t, err, "Couldn't read %s", e)
	return string(ex)
}

func renderToString(t *testing.T, c Chrony) string {
	r, err := RenderChronyTemplate(c)
	require.NoError(t, err, "Could not render chrony configuration")
	return r
}
