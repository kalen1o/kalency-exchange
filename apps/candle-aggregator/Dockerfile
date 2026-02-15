FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/candle-aggregator ./cmd/candle-aggregator

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/candle-aggregator /app/candle-aggregator

EXPOSE 8083
ENTRYPOINT ["/app/candle-aggregator"]
