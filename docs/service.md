# Memos Service Guide

This guide will help you set up Memos as a service on Linux.

⚠ Untested instructions. Please submit a PR if you find any issues.

## Note

To use a privileged port (below 1024), you must set capabilities on the binary:

```sh
sudo setcap 'cap_net_bind_service=+ep' /opt/memos/memos
```

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
Environment=MEMOS_MODE=prod
Environment=MEMOS_PORT=5230
Environment=MEMOS_DATA=/var/opt/memos
RestartSec=5
WorkingDirectory=/var/opt/memos
ExecStart=/var/opt/memos/memos
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

export MEMOS_MODE="prod"
export MEMOS_PORT="5230"
export MEMOS_DATA="/var/opt/memos"

command="/var/opt/memos/memos"
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
```

## Memos configuration

Memos support configuration via environment variables and command line flags. You may set system-wide environment variables, or set them in the service wrapper (recommended).

Some supported environment variables:

```sh
# dev, prod, demo *required*
MEMOS_MODE="prod"

# port to listen on *required*
MEMOS_PORT="5230"

# set addr to 127.0.0.1 to restrict access to localhost
MEMOS_ADDR=""

# data directory: database and asset uploads
MEMOS_DATA="/opt/memos"

# database driver: sqlite, mysql
MEMOS_DRIVER="sqlite"

# database connection string: leave empty for sqlite
# see: https://www.usememos.com/docs/install/database
MEMOS_DSN="dbuser:dbpass@tcp(dbhost)/dbname"

```

See [Memos runtime options](https://www.usememos.com/docs/install/runtime-options) for more details.
