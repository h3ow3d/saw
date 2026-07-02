FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o saw ./cmd/server


FROM alpine:3.21

RUN addgroup -S saw && adduser -S saw -G saw

WORKDIR /app

COPY --from=builder /build/saw        ./saw
COPY --from=builder /build/templates  ./templates
COPY --from=builder /build/static     ./static

USER saw

EXPOSE 8080

CMD ["./saw"]
