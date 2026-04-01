package frr

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/Masterminds/semver/v3"
)

func DetectVersion(log *slog.Logger) (*semver.Version, error) {
	vtysh, err := exec.LookPath("vtysh")
	if err != nil {
		return nil, fmt.Errorf("unable to detect path to vtysh: %w", err)
	}

	// $ vtysh -c "show version"|grep FRRouting
	// FRRouting 10.2.1 (shoot--pz9cjf--mwen-fel-firewall-dcedd) on Linux(6.6.60-060660-generic).

	// $ vtysh -h
	// Usage : vtysh [OPTION...]
	// Integrated shell for FRR (version 10.4.3).

	// Usage : vtysh [OPTION...]
	// Integrated shell for FRR (version 8.4.4).

	c := exec.Command(vtysh, "-h")
	out, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("unable to detect frr version with vtysh output:%s error: %w", string(out), err)
	}

	return parseVersion(log, string(out))
}

func parseVersion(log *slog.Logger, vtyshOutput string) (*semver.Version, error) {
	var frrVersion string

	log.Debug("parseVersion", "vtysh output", vtyshOutput)
	for line := range strings.SplitSeq(vtyshOutput, "\n") {
		if !strings.Contains(line, "Integrated shell for FRR") {
			continue
		}

		_, dirtyVersion, found := strings.Cut(line, "(version ")
		if !found {
			continue
		}

		version, _, found := strings.Cut(dirtyVersion, ").")
		if !found {
			continue
		}

		frrVersion = version
		break
	}

	if frrVersion == "" {
		return nil, fmt.Errorf("unable to detect frr version")
	}

	ver, err := semver.NewVersion(frrVersion)
	if err != nil {
		return nil, fmt.Errorf("unable to parse frr version to semver: %w", err)
	}

	return ver, nil
}
