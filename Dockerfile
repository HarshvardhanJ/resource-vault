FROM golang:alpine AS build

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/archive ./cmd/server

# Final stage
FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S archive && adduser -S archive -G archive \
    && mkdir -p /app/data/storage && chown -R archive:archive /app

COPY --from=build /bin/archive /app/archive
COPY --from=build /app/fixtures /app/fixtures

USER archive:archive

EXPOSE 8080

ENTRYPOINT ["/app/archive"]
