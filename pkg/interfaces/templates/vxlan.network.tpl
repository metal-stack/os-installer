# {{ .Comment }}
[Match]
Name=vni{{ .EVPNIface.VrfID }}

[Link]
MTUBytes=9000

[Network]
Bridge=bridge

[BridgeVLAN]
PVID={{ .EVPNIface.VlanID }}
EgressUntagged={{ .EVPNIface.VlanID }}
