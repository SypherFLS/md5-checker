FROM golang:1.27.3-alpine AS build

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/server ./cmd/api/

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=build /out/server /app/server
COPY config/config.example.yaml /app/config/config.yaml

EXPOSE 8080

CMD ["/app/server"]
