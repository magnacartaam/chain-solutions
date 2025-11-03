FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
#RUN go mod download

COPY . .
COPY vendor ./vendor

RUN CGO_ENABLED=0 GOOS=linux go build -o /main ./cmd/server

FROM scratch
COPY --from=builder /main /main
EXPOSE 8080
CMD ["/main"]
