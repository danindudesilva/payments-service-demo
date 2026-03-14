FROM golang:1.23 AS builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /payments-3ds-demo ./cmd/api

FROM gcr.io/distroless/base-debian12
COPY --from=builder /payments-3ds-demo /payments-3ds-demo
EXPOSE 8080
ENV HTTP_PORT=8080
ENTRYPOINT ["/payments-service-demo"]
