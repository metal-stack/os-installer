# {{ .Comment }}
# network: {{ .EVPNIface.Network }}
[Match]
Name=vrf{{ .EVPNIface.VrfID }}
