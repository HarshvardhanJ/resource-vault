# syntax=docker/dockerfile:1

# Keep this Dockerfile compatible with both the classic Docker builder and
# BuildKit/buildx. The previous version relied on automatic BuildKit
# BUILDPLATFORM/TARGETARCH arguments, which are not defined by the classic
# builder used by some local Docker installations.
FROM golang:1.25-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build for the architecture of the builder. This works natively on local
# Docker installations and on the OCI ARM64 VM. buildx can still build
# separate target-platform images by building this stage for each target.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/archive ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/archive /app/archive
COPY --from=build /src/fixtures /app/fixtures
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/archive"]
