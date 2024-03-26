FROM golang:1.22.1-alpine3.19 AS builder

WORKDIR /app

COPY . .
RUN go mod download
RUN go build -o main /app

FROM alpine:3.19 AS runner

WORKDIR /app

RUN apk update && apk add curl

COPY --from=builder /app/main .

EXPOSE 8000

CMD ["./main"]
