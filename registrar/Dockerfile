ARG alpine_ver=3.12
ARG golang_ver=1.15
ARG VERSION
FROM golang:${golang_ver}-alpine${alpine_ver} AS build
WORKDIR /src/registrard

# hadolint ignore=DL3018
RUN apk add --no-cache make git

# Only download and fetch dependencies when needed
COPY go.mod go.sum Makefile ./
RUN make dep

COPY . .

# Build our application
RUN make build "APP_VERSION=${VERSION}"

FROM alpine:${alpine_ver}

# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates wireguard-tools

# Add our TLS CA
COPY ca.crt /usr/local/share/ca-certificates/registrard-ca.crt
RUN update-ca-certificates

COPY --from=build /src/registrard/bin/* /usr/local/bin/

CMD [ "/usr/local/bin/registrard" ]
