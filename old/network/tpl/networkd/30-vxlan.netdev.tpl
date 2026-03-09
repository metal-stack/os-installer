{{- /*gotype: github.com/metal-stack/os-installer/pkg/network.EVPNIface*/ -}}
{{ .VXLAN.Comment }}
[NetDev]
Name=vni{{ .VXLAN.ID }}
Kind=vxlan

[VXLAN]
VNI={{ .VXLAN.ID }}
Local={{ .VXLAN.TunnelIP }}
UDPChecksum=true
MacLearning=false
DestinationPort=4789
