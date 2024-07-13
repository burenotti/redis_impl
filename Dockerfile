# syntax=docker/dockerfile:1

ARG GO_VERSION=1.22
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS build
WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x

ARG TARGETARCH


RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o /bin/redis ./cmd/redis


FROM gcr.io/distroless/base-debian12:nonroot AS final

COPY --from=build /bin/redis /bin/
COPY config/redis.conf /etc/redis/redis.conf

EXPOSE 6379

ENTRYPOINT [ "/bin/redis", "-config", "/etc/redis/redis.conf" ]
