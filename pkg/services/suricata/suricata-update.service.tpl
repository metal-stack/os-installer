[Unit]
Description=Suricata Intrusion Detection Service Rules Update
After=network-online.target
Wants=network-online.target
Before=suricata.service suricata-ips.service

[Service]
LimitMEMLOCK=infinity
User=root
Group=root
Type=oneshot
ExecStart=/bin/ip vrf exec {{ .DefaultRouteVrf }} /usr/bin/suricata-update

[Install]
WantedBy=multi-user.target
