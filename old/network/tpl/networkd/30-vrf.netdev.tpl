{{- /*gotype: github.com/metal-stack/os-installer/pkg/network.EVPNIface*/ -}}
{{ .VRF.Comment }}
[NetDev]
Name=vrf{{ .VRF.ID }}
Kind=vrf

[VRF]
Table={{ .VRF.Table }}