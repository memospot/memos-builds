# su-exec v0.3

switch user and group id, setgroups and exec

## Building su-exec for ARMv5

### Clone the su-exec repository

```bash
git clone https://github.com/ncopa/su-exec.git
cd su-exec
```

### Enable cross-platform builds with Docker+QEMU

```bash
docker run --privileged --rm tonistiigi/binfmt --install all
```

### Pull the ARMv5 uClibc image

```bash
docker pull dockcross/linux-armv5-uclibc
```

### Generate the dockcross script for this toolchain

```bash
docker run --rm dockcross/linux-armv5-uclibc > ./dockcross-armv5-uclibc
chmod +x ./dockcross-armv5-uclibc
```

### Build

```bash
./dockcross-armv5-uclibc bash -c '$CC -static -Os -s su-exec.c -o su-exec'
```
