FROM golang:1.27.1-alpine AS builder

WORKDIR /app

COPY proto/go.mod proto/go.sum ./proto/
COPY shared/go.mod shared/go.sum ./shared/
COPY spot-service/go.mod spot-service/go.sum ./spot-service/

ENV GOWORK=off
RUN cd spot-service &&  go mod download && go mod tidy

COPY proto/ ./proto/
COPY shared/ ./shared/
COPY spot-service/ ./spot-service/

RUN sh -c 'cd spot-service && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/spot-service ./cmd'

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/spot-service /app/spot-service
COPY spot-service/config.yaml /app/config.yaml

EXPOSE 50052

ENTRYPOINT ["/app/spot-service"]