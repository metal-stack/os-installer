package oscommon

import (
	"log/slog"
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "github.com/metal-stack/os-installer/api/v1"
	"github.com/metal-stack/os-installer/pkg/exec"
	"github.com/metal-stack/os-installer/pkg/test"
	"github.com/spf13/afero"
)

const (
	sampleBlkidOutput = `/dev/sda1: UUID="42d10089-ee1e-0399-445e-755062b63ec8" UUID_SUB="cc57c456-0b2f-6345-c597-d861cc6dd8ac" LABEL="any:0" TYPE="linux_raid_member" PARTLABEL="efi" PARTUUID="273985c8-d097-4123-bcd0-80b4e4e14728"
/dev/sda2: UUID="543eb7f8-98d4-d986-e669-824dbebe69e5" UUID_SUB="54748c60-b566-f391-142c-fb78bb1fc6a9" LABEL="any:1" TYPE="linux_raid_member" PARTLABEL="root" PARTUUID="d7863f4e-af7c-47fc-8c03-6ecdc69bc72d"
/dev/sda3: UUID="fc32a6f0-ee40-d9db-87c8-c9f3a8400c8b" UUID_SUB="582e9b4f-f191-e01e-85fd-2f7d969fbef6" LABEL="any:2" TYPE="linux_raid_member" PARTLABEL="varlib" PARTUUID="e8b44f09-b7f7-4e0d-a7c3-d909617d1f05"
/dev/sdb1: UUID="42d10089-ee1e-0399-445e-755062b63ec8" UUID_SUB="61bd5d8b-1bb8-673b-9e61-8c28dccc3812" LABEL="any:0" TYPE="linux_raid_member" PARTLABEL="efi" PARTUUID="13a4c568-57b0-4259-9927-9ac023aaa5f0"
/dev/sdb2: UUID="543eb7f8-98d4-d986-e669-824dbebe69e5" UUID_SUB="e7d01e93-9340-5b90-68f8-d8f815595132" LABEL="any:1" TYPE="linux_raid_member" PARTLABEL="root" PARTUUID="ab11cd86-37b8-4bae-81e5-21fe0a9c9ae0"
/dev/sdb3: UUID="fc32a6f0-ee40-d9db-87c8-c9f3a8400c8b" UUID_SUB="764217ad-1591-a83a-c799-23397f968729" LABEL="any:2" TYPE="linux_raid_member" PARTLABEL="varlib" PARTUUID="9afbf9c1-b2ba-4b46-8db1-e802d26c93b6"
/dev/md1: LABEL="root" UUID="ace079b5-06be-4429-bbf0-081ea4d7d0d9" TYPE="ext4"
/dev/md0: LABEL="efi" UUID="C236-297F" TYPE="vfat"
/dev/md2: LABEL="varlib" UUID="385e8e8e-dbfd-481e-93a4-cba7f4d5fa02" TYPE="ext4"`
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

func TestDefaultOS_findMDUUID(t *testing.T) {
	tests := []struct {
		name      string
		details   *v1.MachineDetails
		execMocks []test.FakeExecParams
		want      string
		wantFound bool
		wantErr   error
	}{
		{
			name: "no raid",
			details: &v1.MachineDetails{
				RaidEnabled: false,
			},
			want:      "",
			wantFound: false,
			wantErr:   nil,
		},
		{
			name: "with raid",
			details: &v1.MachineDetails{
				RootUUID:    "ace079b5-06be-4429-bbf0-081ea4d7d0d9",
				RaidEnabled: true,
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
			want:      "543eb7f8:98d4d986:e669824d:bebe69e5",
			wantFound: true,
			wantErr:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := slog.Default()

			d := New(&Config{
				Log: log,
				Fs: &afero.Afero{
					Fs: afero.NewMemMapFs(),
				},
				Exec:           exec.New(log).WithCommandFn(test.FakeCmd(t, tt.execMocks...)),
				MachineDetails: tt.details,
			})

			got, gotFound, gotErr := d.findMDUUID(t.Context())
			if diff := cmp.Diff(tt.wantErr, gotErr, test.ErrorStringComparer()); diff != "" {
				t.Errorf("error diff (+got -want):\n%s", diff)
			}

			if tt.wantErr != nil {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (+got -want):\n %s", diff)
			}

			if diff := cmp.Diff(tt.wantFound, gotFound); diff != "" {
				t.Errorf("diff (+got -want):\n %s", diff)
			}
		})
	}
}
