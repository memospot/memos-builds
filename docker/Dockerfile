# Target architectures are limited by the image and its version.

FROM --platform=${BUILDPLATFORM:-linux/amd64} gcr.io/distroless/static:latest as distroless
FROM --platform=${BUILDPLATFORM:-linux/amd64} i386/alpine:latest as target-386
FROM --platform=${BUILDPLATFORM:-linux/amd64} amd64/alpine:latest as target-amd64
FROM --platform=${BUILDPLATFORM:-linux/amd64} amd64/alpine:latest as target-amd64v2
FROM --platform=${BUILDPLATFORM:-linux/amd64} amd64/alpine:latest as target-amd64v3
FROM --platform=${BUILDPLATFORM:-linux/amd64} arm32v5/busybox:1.34-uclibc as target-arm32v5
FROM --platform=${BUILDPLATFORM:-linux/amd64} arm32v6/alpine:latest as target-arm32v6
FROM --platform=${BUILDPLATFORM:-linux/amd64} arm32v7/alpine:latest as target-arm32v7
FROM --platform=${BUILDPLATFORM:-linux/amd64} arm64v8/alpine:latest as target-arm64
FROM --platform=${BUILDPLATFORM:-linux/amd64} ppc64le/alpine:latest as target-ppc64le
FROM --platform=${BUILDPLATFORM:-linux/amd64} riscv64/alpine:edge as target-riscv64
FROM --platform=${BUILDPLATFORM:-linux/amd64} s390x/alpine:latest as target-s390x

# Select appropriate base image.
FROM target-${TARGETARCH}${TARGETVARIANT}

LABEL org.opencontainers.image.title="Memos"
LABEL org.opencontainers.image.description="A privacy-first, lightweight note-taking service."
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.source="https://github.com/lincolnthalles/memos-builds"
LABEL org.opencontainers.image.url="https://usememos.com"
LABEL org.opencontainers.image.vendor="Lincoln Nogueira"

# Setup container OS.

# Copy tzdata and ca-certificates from Google's distroless image.
# This is necessary because the BusyBox image doesn't have tzdata, and also avoid
# running `apk add --no-cache tzdata` in every Alpine build.
COPY --from=distroless /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=distroless /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ARG TARGETPLATFORM TARGETARCH TARGETVARIANT
RUN <<SH
	set -eu
	if ! [ -f /etc/os-release ]; then
		busybox=$(ls --help 2>&1 | head -1)
		if [ "$(echo $busybox | cut -d' ' -f1)" = "BusyBox" ]; then
			echo "PRETTY_NAME=\"$busybox\"" > /etc/os-release
		fi
	fi

	RELEASE=$(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2 | cut -d' ' -f1);
	case $RELEASE in
		Alpine)
			apk upgrade --no-cache
			;;
		Debian)
			mkdir -p /etc/initramfs-tools/
			echo 'update_initramfs=no' > /etc/initramfs-tools/update-initramfs.conf
			rm -f /var/lib/man-db/auto-update

			apt-get -y update
			apt-get -y upgrade
			apt-get --no-install-recommends -y install wget
			apt-get clean
			rm -rf /var/lib/apt/lists/*
			;;
		BusyBox)
			;;
		*)
			echo "Unsupported image: $RELEASE"
			exit 1
			;;
	esac

	mkdir -p /opt/memos
	echo "TARGETPLATFORM=${TARGETPLATFORM}" >> /opt/memos/buildinfo
SH

# Find appropriate Memos binary and copy it.
COPY build/backend/memos_linux_${TARGETARCH}${TARGETVARIANT}/memos /opt/memos/

# Copy front-end files.
COPY build/frontend/dist /opt/memos/dist/

# Copy entrypoint script.
COPY docker/entrypoint.sh /opt/memos/

# Directory to store the data, which can be referenced as the mounting point.
RUN mkdir -p /var/opt/memos
VOLUME /var/opt/memos

ENV TZ=UTC
ENV MEMOS_MODE="prod"
ENV MEMOS_PORT="5230"
ENV MEMOS_METRIC="false"
EXPOSE 5230

HEALTHCHECK --interval=5m --timeout=10s \
	CMD wget -qO - http://localhost:${MEMOS_PORT}/api/v1/ping || exit 1

WORKDIR /opt/memos
ENTRYPOINT [ "sh", "./entrypoint.sh" ]
