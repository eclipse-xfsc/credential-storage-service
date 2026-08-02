FROM golang:1.23.8-alpine3.20 AS build

RUN apk add --no-cache ca-certificates git
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod -o /microservice .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=build /microservice /app/microservice
COPY config.yaml .

EXPOSE 8080
CMD ["/app/microservice"]
