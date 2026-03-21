FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/server /app/server
COPY --from=builder /app/web /app/web
EXPOSE 3000
CMD ["/app/server"]
