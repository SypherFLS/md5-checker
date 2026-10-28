FROM golang:1.27.3

WORKDIR /app/

COPY . .

RUN go build -o app ./cmd/api/

CMD ["./app"] 
