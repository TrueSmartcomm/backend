FROM golang:1.25 as builder

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -v -o app ./cmd/app


FROM alpine:latest

WORKDIR /root/

RUN apk upgrade --no-cache && apk add libc6-compat curl

RUN mkdir -p ./migrations

# Копируем только бинарник в корень
COPY --from=builder /app/app /root/app
COPY --from=builder /go/bin/goose /bin/goose
COPY --from=builder /app/migrations ./migrations

# Делаем бинарник исполняемым (на всякий случай)
RUN chmod +x /root/app

CMD ["./app"]