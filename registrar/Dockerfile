ARG alpine_ver=3.12
ARG golang_ver=1.14
FROM golang:${golang_ver}-alpine${alpine_ver} AS build
WORKDIR /src/registrard

# hadolint ignore=DL3018
RUN apk add --no-cache make git

# Only download and fetch dependencies when needed
COPY go.mod go.sum Makefile ./
RUN make dep

COPY . .
RUN make gobuild

FROM alpine:${alpine_ver}

# hadolint ignore=DL3018
RUN apk add --no-cache ca-certificates

COPY --from=build /src/registrard/bin/* /usr/local/bin/

ENTRYPOINT [ "/usr/local/bin/registrard" ]