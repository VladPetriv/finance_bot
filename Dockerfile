FROM golang:1.23 AS build

WORKDIR /go/build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o finance_bot ./cmd/main.go

##############################################

FROM alpine:3.11
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /go/build/finance_bot .

ENV FB_TELEGRAM_BOT_TOKEN=""
ENV FB_TELEGRAM_WEBHOOK_URL=""
ENV FB_TELEGRAM_SERVER_ADDRESS=""
ENV FB_TELEGRAM_UPDATES_TYPE=""
ENV FB_MONGODB_URI=""
ENV FB_MONGODB_DATABASE="api"
ENV FB_LOGGER_LOG_LEVEL="debug"
ENV FB_LOGGER_LOG_FILENAME=""
ENV FB_LOGGER_PRETTY_LOG_OUTPUT="false"

EXPOSE 8443
EXPOSE 5432

ENTRYPOINT ["./finance_bot"]
