# Memos Linux Service Guide

This guide will help you set up Memos as a service on Linux.

For Windows, see [service-windows.md](service-windows.md).

## Notes

This guide makes the following assumptions:

- `memos` exists in `/usr/local/bin` directory.
- Memos is configured to store its data in `/var/opt/memos` directory.

## systemd

```sh
# Create a data directory
sudo mkdir -p /var/opt/memos

sudo addgroup --system memos
sudo adduser --system --disabled-login --disabled-password --shell /sbin/nologin --group memos

sudo chown -R memos:memos /var/opt/memos

# Create a systemd service file
sudo tee /etc/systemd/system/memos.service <<EOF
[Unit]
Description=Memos Service
Wants=network-online.target
After=network-online.target

[Service]
Type=notify
Environment=MEMOS_PORT=5230
Environment=MEMOS_DATA=/var/opt/memos
RestartSec=5
WorkingDirectory=/var/opt/memos
ExecStart=/usr/local/bin/memos
Restart=on-failure
User=memos
Group=memos

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl start memos
sudo systemctl enable memos
```

## openrc

```sh
# Create a data directory
mkdir -p /var/opt/memos

addgroup -S memos
adduser -S -D -h /var/opt/memos -s /sbin/nologin memos -G memos

chown -R memos:memos /var/opt/memos

# Create an openrc service file
tee /etc/init.d/memos <<EOF
#!/sbin/openrc-run

name="Memos Service"
description="Memos Service"

export MEMOS_PORT="5230"
export MEMOS_DATA="/var/opt/memos"

command="/usr/local/bin/memos"
command_user="memos"
command_background="yes"
command_args=""

pidfile="/run/memos.pid"
respawn_delay="5"
respawn_max="0"
start_stop_daemon_args="--background --make-pidfile --pidfile /run/memos.pid --chuid memos:memos"

depend() {
 after network-online
 use network-online
}

EOF

chmod +x /etc/init.d/memos

# start on boot
rc-update add memos

# start now
rc-service memos start

# debugging
rc-service memos status
cat /var/log/messages | grep memos
journalctl /usr/local/bin/memos
```

## Binding on privileged ports

To use a privileged port (below 1024), you must set capabilities on the binary:

```sh
sudo setcap 'cap_net_bind_service=+ep' /usr/local/bin/memos
```

> [!IMPORTANT]
> You probably don't want to bind on a privileged port, like 80 or 443, as Memos doesn't have built-in TLS encryption.
> It's best to bind Memos to 5230 and use a reverse proxy, like [Caddy](https://caddyserver.com/), so you'll also have TLS encryption, provided you have a domain name.

## Memos configuration

See [configuration.md](configuration.md).

## Next steps

Set up a reverse proxy, like [Caddy](https://caddyserver.com/), to add TLS encryption, provided you have a domain name.
