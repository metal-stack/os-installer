package test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	var f FakeExecParams
	err := json.Unmarshal([]byte(os.Args[3]), &f)
	require.NoError(t, err)

	_, err = fmt.Fprint(os.Stdout, f.Output)
	require.NoError(t, err)

	os.Exit(f.ExitCode)
}
