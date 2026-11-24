[Unit]
Description=Suricata Intrusion Prevention Service (nfqueue)
Documentation=man:suricata(8) man:suricatasc(8)
After=network.target network-online.target suricata-update.service
Wants=network-online.target suricata-update.service
Conflicts=suricata.service

[Service]
Type=simple
EnvironmentFile=-/etc/default/suricata
ExecStartPre=/bin/rm -f $PIDFILE
ExecStart=/usr/bin/suricata -c $SURCONF --pidfile $PIDFILE -q 0
ExecReload=/usr/bin/suricatasc -c reload-rules
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
