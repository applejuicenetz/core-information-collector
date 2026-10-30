FROM docker.io/library/golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -tags nogui -trimpath -ldflags "-s -w -X github.com/applejuicenetz/core-information-collector/internal/buildinfo.Version=${VERSION}" -o /ajcollector ./cmd/collector

FROM docker.io/library/alpine:3
RUN mkdir -p /app/appleJuice/collector
COPY --from=build /ajcollector /usr/local/bin/ajcollector
ENV AJ_NOTRAY=1 \
    AJ_COLLECTOR_CONFIG_DIR=/app/appleJuice/collector
WORKDIR /app
CMD ["ajcollector"]
