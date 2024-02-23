FROM docker.io/golang:1.22 as builder
WORKDIR /app
COPY go.mod go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -a \
    -ldflags="-s -w" \
    -o ./api .

FROM scratch as runtime
WORKDIR /app
COPY --from=builder /app/api .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 5000
CMD ["./api"]