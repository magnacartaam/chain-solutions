FROM golang:1.25-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .

FROM base AS dev
RUN go install github.com/air-verse/air@latest
CMD ["air"]


FROM base AS builder
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /main ./cmd/server


FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /main .
EXPOSE 8080
CMD ["./main"]