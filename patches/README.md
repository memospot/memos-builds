# Patches

Any patch files put in here will be applied to the source code using `git apply`,
with fallback to `patch`.

> [!NOTE]
> The modernc.org/sqlite patching is now done programatically.
> The logic can be found in [`patch.go`](../.dagger/patch.go).
