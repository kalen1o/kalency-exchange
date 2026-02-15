FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/ledger-writer ./cmd/ledger-writer

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/ledger-writer /app/ledger-writer

EXPOSE 8084
ENTRYPOINT ["/app/ledger-writer"]
