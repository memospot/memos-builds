# Memos server configuration

Memos support configuration via environment variables and command line flags.

You may set system-wide environment variables, set them in the service wrapper (recommended), or pass them at the process command line.

## Environment variables

Some supported environment variables:

```sh
# dev, prod, demo *required* before v0.26.0.
# Starting from v0.26.0, this variable is retired and the database is always `prod` unless the new variable `MEMOS_DEMO=true` is set.
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
# see: https://usememos.com/docs/configuration/database
MEMOS_DSN="dbuser:dbpass@tcp(dbhost)/dbname"
```

## Details

See [Memos configuration options](https://usememos.com/docs/configuration).

## Extra

For all supported environment variables that may be left out of the documentation, see [cmd/memos.go](https://github.com/usememos/memos/blob/main/cmd/memos/main.go#99). All bound flags in `init()` are also supported as uppercased environment variables, prefixed with `MEMOS_`.
