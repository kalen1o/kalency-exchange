FROM golang:1.25-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/matching-engine ./cmd/matching-engine

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/matching-engine /app/matching-engine

EXPOSE 8081
ENTRYPOINT ["/app/matching-engine"]
