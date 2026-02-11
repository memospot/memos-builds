# Running Memos server on Android

## Setup Termux

### 1. Install [Termux](https://play.google.com/store/apps/details?id=com.termux)

### 2. Launch Termux and set up [Termux storages](https://wiki.termux.com/wiki/Termux-setup-storage)

> [!TIP]
> This will prompt you for storage access permission.

```bash
termux-setup-storages
```

## Manual installation

1. Check your device CPU architecture with `uname -m`.

2. Download a `linux` release that matches your CPU architecture from <https://github.com/memospot/memos-builds/releases>. `curl -L -O "$URL"`

3. Extract the release to a directory, e.g. `~/memos`. `tar -xf "$FILENAME"`

4. Copy the binary to the Termux home directory: `cp ~/storage/shared/memos .`

5. Make it executable: `chmod +x ./memos`

6. Run Memos: `MEMOS_DATA=. MEMOS_PORT=5230 ./memos`

7. Memos will be available at `http://localhost:5230` and on local network at `http://<device-ip>:5230`

## Memos configuration

See [configuration.md](configuration.md).

> [!WARNING]
> ⚠ As stated at [Termux Wiki](https://wiki.termux.com/wiki/Internal_and_external_storage), all data under Termux's home directory will be deleted if you uninstall the app.
