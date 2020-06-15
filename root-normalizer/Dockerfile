FROM alpine:3.12

RUN apk add --no-cache bash socat

COPY ./init.sh /init.sh

ENTRYPOINT ["/init.sh"]
