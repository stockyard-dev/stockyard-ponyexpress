FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/ponyexpress ./cmd/ponyexpress/
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata curl
COPY --from=builder /bin/ponyexpress /usr/local/bin/ponyexpress
ENV PORT="9010" DATA_DIR="/data"
EXPOSE 9010
HEALTHCHECK --interval=30s --timeout=5s CMD curl -sf http://localhost:9010/health || exit 1
ENTRYPOINT ["ponyexpress"]
