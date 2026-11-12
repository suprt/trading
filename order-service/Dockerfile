FROM golang:1.27.1-alpine AS builder

WORKDIR /app

COPY proto/go.mod proto/go.sum ./proto/
COPY shared/go.mod shared/go.sum ./shared/
COPY order-service/go.mod order-service/go.sum ./order-service/

ENV GOWORK=off
RUN cd order-service &&  go mod download && go mod tidy

COPY proto/ ./proto/
COPY shared/ ./shared/
COPY order-service/ ./order-service/

RUN sh -c 'cd order-service && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/order-service ./cmd'

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/order-service /app/order-service
COPY order-service/config.yaml /app/config.yaml

EXPOSE 50053

ENTRYPOINT ["/app/order-service"]