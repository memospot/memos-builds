# 	The `--platform` flag is used to specify the target platform in a format supported by the *base image*.
#
# 	Platform format is not consistent across image distros and versions (e.g. arm/arm32),
# and when omitted, it will use whatever is supplied to the `docker buildx build` command,
# and the build may fail if we are not using the same format as the base image.
#
# 	By specifying the platform flag, it's possible to compatibilize different images without changing
# the build command, which would imply changing the published platform formats for this image.
FROM --platform=${BUILDPLATFORM:-linux/amd64} 	gcr.io/distroless/static:latest AS distroless
FROM 											alpine:3.21 					AS target-386
FROM  											alpine:3.21 					AS target-amd64
FROM --platform=amd64  							alpine:3.21 					AS target-amd64v2
FROM --platform=amd64							alpine:3.21 					AS target-amd64v3
FROM --platform=linux/arm/v5 					busybox:1.36.1-uclibc 			AS target-armv5
FROM --platform=linux/arm/v6 					alpine:3.21  					AS target-armv6
FROM --platform=linux/arm/v7 					alpine:3.21  					AS target-armv7
FROM  											alpine:3.21  					AS target-arm64
FROM  											alpine:3.21  					AS target-ppc64le
FROM  											alpine:3.21 					AS target-riscv64
FROM  											alpine:3.21  					AS target-s390x

# Select appropriate base image and setup container OS.
FROM target-${TARGETARCH}${TARGETVARIANT}

LABEL org.opencontainers.image.title="Memos"
LABEL org.opencontainers.image.description="A privacy-first, lightweight note-taking service."
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.vendor="Lincoln Nogueira"
LABEL org.opencontainers.image.source="https://github.com/memospot/memos-builds"
LABEL org.opencontainers.image.url="https://usememos.com"

# Copy tzdata and ca-certificates from Google's distroless image.
# This is necessary because the BusyBox image doesn't have tzdata, and also avoid
# running `apk add --no-cache tzdata` in every Alpine build.
COPY --from=distroless /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=distroless /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ARG TARGETPLATFORM TARGETARCH TARGETVARIANT
SHELL [ "/bin/ash", "-euo", "pipefail", "-c" ]
RUN <<SH
    if ! [ -f /etc/os-release ]; then
        # Setup os-release file for BusyBox.
        busybox=$(find --help 2>&1 | head -1)
        if [ "$(echo "$busybox" | cut -d' ' -f1)" = "BusyBox" ]; then
            echo "PRETTY_NAME=\"$busybox\"" > /etc/os-release
        fi
    fi

    # Run distro-specific commands.
    RELEASE=$(grep PRETTY_NAME < /etc/os-release | cut -d'"' -f2 | cut -d' ' -f1);
    case $RELEASE in
        Alpine)
            apk upgrade --no-cache
            ;;
        *)
            echo "No action needed for distro ${RELEASE}."
            ;;
    esac

    mkdir -p /opt/memos
    echo "TARGETPLATFORM=${TARGETPLATFORM}" >> /opt/memos/buildinfo
SH

# Find appropriate Memos binary and copy it.
# Binaries are built with goreleaser and then renamed upfront to match this format.
COPY build/memos_linux_${TARGETARCH}${TARGETVARIANT}/memos /opt/memos/

# Copy front-end files.
# COPY memos/server/router/frontend/dist /opt/memos/dist/

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
    CMD wget -qO - http://localhost:${MEMOS_PORT}/healthz || exit 1

WORKDIR /opt/memos
ENTRYPOINT ["ash", "./entrypoint.sh", "./memos"]
