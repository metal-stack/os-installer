package ubuntu_test

import (
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/metal-stack/os-installer/pkg/exec"
	oscommon "github.com/metal-stack/os-installer/pkg/installer/os/common"
	"github.com/metal-stack/os-installer/pkg/installer/os/ubuntu"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
)

const (
	sampleMdadmDetailOutput = `MD_LEVEL=raid1
MD_DEVICES=2
MD_METADATA=1.0
MD_UUID=543eb7f8:98d4d986:e669824d:bebe69e5
MD_DEVNAME=1
MD_NAME=any:1
MD_DEVICE_dev_sdb2_ROLE=1
MD_DEVICE_dev_sdb2_DEV=/dev/sdb2
MD_DEVICE_dev_sda2_ROLE=0
MD_DEVICE_dev_sda2_DEV=/dev/sda2`
)

func Test_os_BuildCMDLine(t *testing.T) {
	tests := []struct {
		name      string
		details   *v1.MachineDetails
		execMocks []test.FakeExecParams
		want      string
		wantErr   error
	}{
		{
			name: "no raid",
			details: &v1.MachineDetails{
				RootUUID:    "543eb7f8-98d4-d986-e669-824dbebe69e5",
				RaidEnabled: false,
				Console:     "ttyS1,115200n8",
			},
			want:    "console=ttyS1,115200n8 root=UUID=543eb7f8-98d4-d986-e669-824dbebe69e5 init=/sbin/init net.ifnames=0 biosdevname=0 nvme_core.io_timeout=300",
			wantErr: nil,
		},
		{
			name: "with raid",
			details: &v1.MachineDetails{
				RootUUID:    "ace079b5-06be-4429-bbf0-081ea4d7d0d9",
				RaidEnabled: true,
				Console:     "ttyS1,115200n8",
			},
			execMocks: []test.FakeExecParams{
				{
					WantCmd:  []string{"blkid"},
					Output:   sampleBlkidOutput,
					ExitCode: 0,
				},
				{
					WantCmd:  []string{"mdadm", "--detail", "--export", "/dev/md1"},
					Output:   sampleMdadmDetailOutput,
					ExitCode: 0,
				},
			},
			want:    "console=ttyS1,115200n8 root=UUID=ace079b5-06be-4429-bbf0-081ea4d7d0d9 init=/sbin/init net.ifnames=0 biosdevname=0 nvme_core.io_timeout=300 rdloaddriver=raid0 rdloaddriver=raid1 rd.md.uuid=543eb7f8:98d4d986:e669824d:bebe69e5",
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.Default()

			d := ubuntu.New(&oscommon.Config{
				Log: log,
				Fs: &afero.Afero{
					Fs: afero.NewMemMapFs(),
				},
				Exec:           exec.New(log).WithCommandFn(test.FakeCmd(t, tt.execMocks...)),
				MachineDetails: tt.details,
			})

			got, gotErr := d.BuildCMDLine(t.Context())
			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (+got -want):\n %s", diff)
			}
		})
	}
}
