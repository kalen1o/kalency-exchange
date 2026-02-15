FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/market-sim ./cmd/market-sim

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/market-sim /app/market-sim

EXPOSE 8082
ENTRYPOINT ["/app/market-sim"]
