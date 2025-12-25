# syntax=docker/dockerfile:1
FROM golang:1.25.5-alpine AS build

RUN apk add --no-cache ca-certificates
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/user ./cmd/user

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/user /app/user

EXPOSE 9090 8080

ENTRYPOINT ["/app/user"]
