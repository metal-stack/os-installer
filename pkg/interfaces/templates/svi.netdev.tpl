# {{ .Comment }}
[NetDev]
Name=vlan{{ .EVPNIface.VrfID }}
Kind=vlan

[VLAN]
Id={{ .EVPNIface.VlanID }}
