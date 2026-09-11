# syntax=docker/dockerfile:1

# Compatible with the classic Docker builder as well as BuildKit/buildx.
FROM golang:1.25-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/archive ./cmd/server

# Deliberately distroless and non-root for the runtime.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=build --chown=nonroot:nonroot /out/archive /app/archive
COPY --from=build --chown=nonroot:nonroot /src/fixtures /app/fixtures

# Runtime storage is supplied by the Docker volume in compose.yaml.
# Do not add RUN commands here: distroless images do not contain /bin/sh.

USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/archive"]
