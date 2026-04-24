FROM golang:1.22-alpine AS builder
WORKDIR /build
COPY src/ .
RUN go build -o heimdall .

FROM alpine:3.19
WORKDIR /app
RUN apk add --no-cache bash openssh-client
COPY --from=builder /build/heimdall /usr/local/bin/heimdall
ENTRYPOINT ["heimdall"]
