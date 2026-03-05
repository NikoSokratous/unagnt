# Match go.mod (go 1.25.0); use 1.24 if 1.25 image unavailable
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /unagnt ./cmd/unagnt
RUN go build -o /unagntd ./cmd/unagntd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /unagnt /usr/local/bin/
COPY --from=builder /unagntd /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/unagntd"]
