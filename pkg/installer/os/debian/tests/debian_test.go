package debian_test

import (
	"encoding/json"
	"fmt"
	goos "os"
	"testing"

	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/stretchr/testify/require"
)

func TestHelperProcess(t *testing.T) {
	if goos.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	var f test.FakeExecParams
	err := json.Unmarshal([]byte(goos.Args[3]), &f)
	require.NoError(t, err)

	_, err = fmt.Fprint(goos.Stdout, f.Output)
	require.NoError(t, err)

	goos.Exit(f.ExitCode)
}
