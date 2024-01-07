FROM alpine:latest AS monolithic

RUN apk add --no-cache tzdata
ENV TZ="UTC"

COPY build/dist /usr/local/memos/dist
COPY build/backend/memos /usr/local/memos/

# Directory to store the data, which can be referenced as the mounting point.
RUN mkdir -p /var/opt/memos
VOLUME /var/opt/memos

ENV MEMOS_MODE="prod"
ENV MEMOS_PORT="5230"
EXPOSE 5230

WORKDIR /usr/local/memos
ENTRYPOINT ["./memos"]
