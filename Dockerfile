FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./

# Proxy for OrbStack (host.docker.internal reaches the macOS host)
ARG HTTP_PROXY
ARG HTTPS_PROXY
ENV HTTP_PROXY=${HTTP_PROXY}
ENV HTTPS_PROXY=${HTTPS_PROXY}
ENV GOPROXY=https://proxy.golang.org,direct

RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /server /server
COPY migrations/ /migrations/

EXPOSE 8080

ENTRYPOINT ["/server"]
