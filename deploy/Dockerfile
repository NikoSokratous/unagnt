FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /agentctl ./cmd/agentctl
RUN go build -o /agentd ./cmd/agentd

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /agentctl /usr/local/bin/
COPY --from=builder /agentd /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/agentd"]
