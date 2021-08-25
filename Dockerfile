FROM golang:1.16.2 as builder

WORKDIR /build

COPY go.mod go.sum ./
COPY internal/ ./internal
COPY cmd/ ./cmd

ENV GOPRIVATE=pkg.aiocean.dev/*,github.com/aiocean/*
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server

FROM scratch
WORKDIR /root/

COPY --from=builder /build/server /server
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/server"]
