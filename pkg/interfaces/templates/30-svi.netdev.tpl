{{- /*gotype: github.com/metal-stack/os-installer/pkg/network.EVPNIface*/ -}}
{{ .SVI.Comment }}
[NetDev]
Name=vlan{{ .VRF.ID }}
Kind=vlan

[VLAN]
Id={{ .SVI.VLANID }}
