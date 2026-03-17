package operatingsystem_test

import (
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	apiv2 "github.com/metal-stack/api/go/metalstack/api/v2"
	v1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/metal-stack/os-installer/pkg/exec"
	operatingsystem "github.com/metal-stack/os-installer/pkg/installer/os"
	"github.com/metal-stack/os-installer/pkg/installer/os/almalinux"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/metal-stack/os-installer/pkg/installer/os/debian"
	"github.com/metal-stack/os-installer/pkg/installer/os/ubuntu"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	ubuntuRelease = `PRETTY_NAME="Ubuntu 24.04.4 LTS"
NAME="Ubuntu"
VERSION_ID="24.04"
VERSION="24.04.4 LTS (Noble Numbat)"
VERSION_CODENAME=noble
ID=ubuntu
ID_LIKE=debian
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
UBUNTU_CODENAME=noble
LOGO=ubuntu-logo
`
	debianRelease = `PRETTY_NAME="Debian GNU/Linux 13 (trixie)"
NAME="Debian GNU/Linux"
VERSION_ID="13"
VERSION="13 (trixie)"
VERSION_CODENAME=trixie
DEBIAN_VERSION_FULL=13.4
ID=debian
HOME_URL="https://www.debian.org/"
SUPPORT_URL="https://www.debian.org/support"
BUG_REPORT_URL="https://bugs.debian.org/"
`
	almalinuxRelease = `NAME="AlmaLinux"
VERSION="10.1 (Heliotrope Lion)"
ID="almalinux"
ID_LIKE="rhel centos fedora"
VERSION_ID="10.1"
PLATFORM_ID="platform:el10"
PRETTY_NAME="AlmaLinux 10.1 (Heliotrope Lion)"
ANSI_COLOR="0;34"
LOGO="fedora-logo-icon"
CPE_NAME="cpe:/o:almalinux:almalinux:10.1"
HOME_URL="https://almalinux.org/"
DOCUMENTATION_URL="https://wiki.almalinux.org/"
VENDOR_NAME="AlmaLinux"
VENDOR_URL="https://almalinux.org/"
BUG_REPORT_URL="https://bugs.almalinux.org/"

ALMALINUX_MANTISBT_PROJECT="AlmaLinux-10"
ALMALINUX_MANTISBT_PROJECT_VERSION="10.1"
REDHAT_SUPPORT_PRODUCT="AlmaLinux"
REDHAT_SUPPORT_PRODUCT_VERSION="10.1"
SUPPORT_END=2035-06-01
`
	unknownRelease = `NAME="EndeavourOS"
PRETTY_NAME="EndeavourOS"
ID="endeavouros"
ID_LIKE="arch"
BUILD_ID="2025.03.19"
ANSI_COLOR="38;2;23;147;209"
HOME_URL="https://endeavouros.com"
DOCUMENTATION_URL="https://discovery.endeavouros.com"
SUPPORT_URL="https://forum.endeavouros.com"
BUG_REPORT_URL="https://forum.endeavouros.com/c/general-system/endeavouros-installation"
PRIVACY_POLICY_URL="https://endeavouros.com/privacy-policy-2"
LOGO="endeavouros"`
)

func Test_New(t *testing.T) {
	tests := []struct {
		name       string
		explicitOS *string
		fsMocks    func(fs *afero.Afero)
		want       any
		wantErr    error
	}{
		{
			name: "detect ubuntu",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, afero.WriteFile(fs, operatingsystem.OsReleasePath, []byte(ubuntuRelease), 0777))
			},
			want: &ubuntu.Os{},
		},
		{
			name: "detect debian",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, afero.WriteFile(fs, operatingsystem.OsReleasePath, []byte(debianRelease), 0777))
			},
			want: &debian.Os{},
		},
		{
			name: "detect almalinux",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, afero.WriteFile(fs, operatingsystem.OsReleasePath, []byte(almalinuxRelease), 0777))
			},
			want: &almalinux.Os{},
		},
		{
			name: "detect default for unknown",
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, afero.WriteFile(fs, operatingsystem.OsReleasePath, []byte(unknownRelease), 0777))
			},
			want: &oscommon.DefaultOS{},
		},
		{
			name:       "explicitly want almalinux impl on unknown os",
			explicitOS: new("almalinux"),
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, afero.WriteFile(fs, operatingsystem.OsReleasePath, []byte(unknownRelease), 0777))
			},
			want: &almalinux.Os{},
		},
		{
			name:       "explicitly want unsupported",
			explicitOS: new("foo"),
			fsMocks: func(fs *afero.Afero) {
				require.NoError(t, afero.WriteFile(fs, operatingsystem.OsReleasePath, []byte(unknownRelease), 0777))
			},
			wantErr: fmt.Errorf(`os with name "foo" is not supported`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				log = slog.Default()
				fs  = &afero.Afero{
					Fs: afero.NewMemMapFs(),
				}
			)

			if tt.fsMocks != nil {
				tt.fsMocks(fs)
			}

			os, gotErr := operatingsystem.New(&oscommon.Config{
				Log:            log,
				Name:           tt.explicitOS,
				Fs:             fs,
				Exec:           exec.New(log).WithCommandFn(test.FakeCmd(t)),
				MachineDetails: &v1.MachineDetails{},
				Allocation:     &apiv2.MachineAllocation{},
			})
			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			assert.IsType(t, tt.want, os)
		})
	}
}
