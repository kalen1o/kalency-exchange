FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/gateway-api ./cmd/gateway-api

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/gateway-api /app/gateway-api

EXPOSE 8080
ENTRYPOINT ["/app/gateway-api"]
