FROM golang:1.16-alpine AS builder
RUN apk update && \
    apk add --no-cache git gcc musl-dev
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o /out/bot
COPY ./www /out/www

FROM alpine:latest
WORKDIR /opt/bot
COPY --from=builder /out/ /opt/bot/
CMD /opt/bot/bot
