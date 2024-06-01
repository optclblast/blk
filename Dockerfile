FROM golang:alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0

RUN apk update --no-cache && apk add --no-cache tzdata

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/blk cmd/app/main.go

FROM alpine:latest 
WORKDIR /build
COPY --from=builder /app/blk .
RUN chmod +x ./blk

EXPOSE 8080

CMD ["./blk"]
