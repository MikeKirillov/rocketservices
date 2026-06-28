FROM alpine:3.23

ARG MIGRATIONS_DIR

WORKDIR /opt/app

RUN apk add --no-cache ca-certificates \
    && adduser -D -h /opt/app web

ADD https://github.com/pressly/goose/releases/download/v3.14.0/goose_linux_x86_64 /usr/local/bin/goose

COPY ${MIGRATIONS_DIR}/*.sql ./db/migrations/

RUN chmod +x /usr/local/bin/goose \
    && chown -R web:web /opt/app

USER web

ENTRYPOINT ["goose"]