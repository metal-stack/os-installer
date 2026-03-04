{{- /*gotype: github.com/metal-stack/os-installer/pkg/network.EVPNIface*/ -}}
{{ .VRF.Comment }}
[Match]
Name=vrf{{ .VRF.ID }}