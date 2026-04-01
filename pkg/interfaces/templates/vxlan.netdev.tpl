# {{ .Comment }}
# network: {{ .EVPNIface.Network }}
[NetDev]
Name=vni{{ .EVPNIface.VrfID }}
Kind=vxlan

[VXLAN]
VNI={{ .EVPNIface.VrfID }}
Local={{ .UnderlayIP }}
UDPChecksum=true
MacLearning=false
DestinationPort=4789
