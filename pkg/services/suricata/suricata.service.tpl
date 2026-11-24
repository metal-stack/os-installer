[Unit]
Description=Suricata Intrusion Detection Service (af-packet)
Documentation=man:suricata(8) man:suricatasc(8)
After=network.target network-online.target suricata_update.service
Wants=network-online.target suricata_update.service
Conflicts=suricata-ips.service

[Service]
Type=simple
EnvironmentFile=-/etc/default/suricata
ExecStartPre=/bin/rm -f $PIDFILE
ExecStart=/usr/bin/suricata -c $SURCONF --pidfile $PIDFILE --af-packet
ExecReload=/usr/bin/suricatasc -c reload-rules
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
