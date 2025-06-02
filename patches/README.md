# Patches

- 0.24.4-sqlite-1.37.1

  Fixes a compilation failure on alternate platforms due to mismatch between the version of `modernc.org/sqlite` and `modernc.org/libc`.
  See:
  - <https://pkg.go.dev/modernc.org/sqlite#hdr-Fragile_modernc_org_libc_dependency>
  - <https://gitlab.com/cznic/sqlite/-/issues/177>
