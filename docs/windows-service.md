# Memos Windows Service Guide

While Memos is designed to run on Docker, you may also run Memos as a Windows service.
It will run under the `SYSTEM` account and start automatically at system boot.

## ⚠ Notes

All service management methods require admin privileges.
For convenience, use [gsudo](https://gerardog.github.io/gsudo/docs/install), or open a new PowerShell terminal as admin:

```powershell
Start-Process powershell -Verb RunAs
```

This guide makes the following assumptions:

- You are using Powershell.
- `memos.exe` exists in `C:\ProgramData\memos` directory.
- Memos is configured to store its data in `C:\ProgramData\memos` directory.

If you want the service to be reachable from the network, you should also configure your firewall to allow inbound connections to the server:

  ```powershell
  # Allow memos.exe on Windows Firewall
  New-NetFirewallRule -DisplayName "Memos" -Direction Inbound -Program "$Env:ProgramData\memos\memos.exe" -Action Allow -Protocol TCP
  ```

## Windows Service Wrappers

Choose one of the following methods to install Memos as a service.

### 1. [NSSM](https://nssm.cc/download)

NSSM is a lightweight service wrapper. It uses very little memory and CPU time, and it is stable and reliable.

The downside is that it doesn't support configuration files, so you have to use the command line to configure the service.

You may download and extract `nssm.exe` in the same directory as `memos.exe`, or add its directory to your system PATH. Prefer the latest 64-bit version of `nssm.exe`.

Also, nssm is available on [Chocolatey](https://chocolatey.org/) and [Scoop](https://scoop.sh/) Windows Package Managers:

```powershell
choco install nssm
scoop install nssm
```

NSSM command line usage:

```powershell
Set-Location -Path "$Env:ProgramData\memos"

# Install memos as a service
nssm install memos "$Env:ProgramData\memos\memos.exe"
nssm set memos DisplayName "Memos Service"
nssm set memos Description "A privacy-first, lightweight note-taking service. https://usememos.com/"

# Configure memos
nssm set memos AppEnvironmentExtra MEMOS_MODE="prod" MEMOS_PORT="5230" MEMOS_DATA="$Env:ProgramData\memos"

# Delay service auto start *optional*
nssm set memos Start SERVICE_DELAYED_AUTO_START

# Edit service using NSSM built-in GUI
nssm edit memos

# Start the service
nssm start memos

# Remove the service, if you ever need to
nssm remove memos confirm
```

### 2. Using [WinSW](https://github.com/winsw/winsw)

Download `WinSW-net461.exe` from [GitHub Releases](https://github.com/winsw/winsw/releases/latest). Then, put it in the same directory as `memos.exe` and rename `WinSW-net461.exe` to `memos-service.exe`.

Now, in the same directory, create a service configuration file named `memos-service.xml`:

```xml
<service>
    <id>memos</id>
    <name>Memos Service</name>
    <description>A privacy-first, lightweight note-taking service. https://usememos.com/</description>
    <onfailure action="restart" delay="10 sec"/>
    <executable>%BASE%\memos.exe</executable>
    <env name="MEMOS_MODE" value="prod" />
    <env name="MEMOS_ADDR" value="" />
    <env name="MEMOS_PORT" value="5230" />
    <env name="MEMOS_DATA" value="%ProgramData%\memos" />
    <env name="MEMOS_METRIC" value="true" />
    <delayedAutoStart>true</delayedAutoStart>
    <log mode="none" />
</service>
```

Then, install the service:

```powershell
Set-Location -Path "$Env:ProgramData\memos"

# Install the service
.\memos-service.exe install

# Start the service
.\memos-service.exe start

# Remove the service, if you ever need to
.\memos-service.exe uninstall
```

### Manage the service

You may use the `net` command to manage the service:

```powershell
net start memos
net stop memos
```

If the service installation was successful, the service would appear in the Windows Services Manager `services.msc` labeled as `Memos Service`.

## Memos configuration

Memos support configuration via environment variables and command line flags. You may set system-wide environment variables, or set them in the service wrapper (recommended).

Currently, Memos supports the following environment variables:

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
# see: https://www.usememos.com/docs/advanced-settings/mysql
MEMOS_DSN="dbuser:dbpass@tcp(dbhost)/dbname"

# allow metric collection
MEMOS_METRIC="true"
```

For all supported environment variables, see [cmd/memos.go](https://github.com/usememos/memos/blob/main/cmd/memos.go#L106). All bound flags in `init()` are also supported as environment variables, prefixed with `MEMOS_`.

To set-up Memos with MySQL, see [Memos MySQL Guide](https://www.usememos.com/docs/advanced-settings/mysql).

## Additional notes

- When `--data` / `MEMOS_DATA` is not specified, Memos will store its data in the following directory: `C:\ProgramData\memos`.

- If the service fails to start, you should inspect the Windows Event Viewer `eventvwr.msc`.

After setup, Memos will be accessible at [http://localhost:5230](http://localhost:5230), if you didn't change the default port.
