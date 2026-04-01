# {{ .Comment }}
[Match]
PermanentMACAddress={{ .Mac }}

[Link]
Name=lan{{ .Index }}
NamePolicy=
MTUBytes={{ .MTU }}
