FROM --platform=$BUILDPLATFORM oven/bun:alpine AS frontend-build
WORKDIR /frontend

COPY ./frontend/package.json ./frontend/bun.lock ./
RUN bun install --frozen-lockfile

COPY ./frontend ./
RUN bun run build

FROM --platform=$BUILDPLATFORM golang:alpine AS build-env
WORKDIR /src

COPY ./go.mod ./go.sum ./
RUN go mod download

COPY . .
COPY --from=frontend-build /frontend/dist ./frontend/dist

ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 GOEXPERIMENT=jsonv2 go build -ldflags "-s -w" -v -o waka3x main.go
# Need a statically linked healthcheck binary because we can't use curl in a distroless image in a straightforward way
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -ldflags "-s -w" -v -o healthcheck scripts/healthcheck.go

WORKDIR /staging
RUN mkdir ./data ./app && \
    cp /src/waka3x app/ && \
    cp /src/healthcheck app/ && \
    cp /src/config.default.yml app/config.yml && \
    sed -i 's/listen_ipv6: ::1/listen_ipv6: "-"/g' app/config.yml

# Run Stage

# When running the application using `docker run`, you can pass environment variables
# to override config values using `-e` syntax.
# Available options can be found in [README.md#-configuration](README.md#-configuration)

# Note on the distroless image:
# we could use `base:nonroot`, which already includes ca-certificates and tz, but that one it actually larger than alpine,
# probably because of glibc, whereas alpine uses musl. The `static:nonroot`, doesn't include any libc implementation, because only meant for true static binaries without cgo, etc.

FROM gcr.io/distroless/static:nonroot
WORKDIR /app

# See README.md and config.default.yml for all config options
ENV ENVIRONMENT=prod \
    WAKA3X_DB_TYPE=sqlite3 \
    WAKA3X_DB_USER='' \
    WAKA3X_DB_PASSWORD='' \
    WAKA3X_DB_HOST='' \
    WAKA3X_DB_NAME=/data/waka3x.db \
    WAKA3X_PASSWORD_SALT='' \
    WAKA3X_LISTEN_IPV4='0.0.0.0' \
    WAKA3X_INSECURE_COOKIES='true' \
    WAKA3X_ALLOW_SIGNUP='true'

COPY --from=build-env --chown=root:root /staging/app /app
COPY --from=build-env --chown=nonroot:nonroot /staging/data /data

LABEL org.opencontainers.image.url="https://github.com/zetkey/waka3x" \
    org.opencontainers.image.documentation="https://github.com/zetkey/waka3x" \
    org.opencontainers.image.source="https://github.com/zetkey/waka3x" \
    org.opencontainers.image.title="waka3x" \
    org.opencontainers.image.licenses="MIT" \
    org.opencontainers.image.description="A minimalist, self-hosted WakaTime-compatible backend for coding statistics"

USER nonroot

EXPOSE 3000

# For long-running migrations, you might want to override `---health-start-period` as part of `docker run` or disable healthchecks entirely with `--no-healtcheck`
HEALTHCHECK --interval=60s --timeout=3s --start-period=120s --retries=3 CMD ["/app/healthcheck"]

ENTRYPOINT ["/app/waka3x"]
