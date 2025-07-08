# Start from official Go image
FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache git && mkdir -p /data/files /data/chunks /data/metadata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server .

EXPOSE 8080

CMD ["./server"]
