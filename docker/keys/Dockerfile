FROM alpine:3.21

RUN apk add --no-cache openssl

COPY scripts/generate-dev-keys.sh /usr/local/bin/generate-dev-keys
RUN chmod +x /usr/local/bin/generate-dev-keys

ENTRYPOINT ["/usr/local/bin/generate-dev-keys", "/keys"]
