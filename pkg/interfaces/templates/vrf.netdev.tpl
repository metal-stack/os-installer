# {{ .Comment }}
# network: {{ .EVPNIface.Network }}
[NetDev]
Name=vrf{{ .EVPNIface.VrfID }}
Kind=vrf

[VRF]
Table={{ .EVPNIface.VlanID }}
