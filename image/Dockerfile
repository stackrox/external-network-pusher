FROM alpine:3.12

# Needed by network-crawler script
RUN apk add curl

COPY network-crawler /usr/bin/network-crawler

ENTRYPOINT ["/usr/bin/network-crawler"]
