FROM golang:1.21.5-alpine AS base
RUN apk add build-base
RUN apk add --no-cache git

FROM base AS pluginBuild
ARG pluginRepoUrl
ARG pluginTag=main

WORKDIR /app

RUN git clone -b $pluginTag $pluginRepoUrl .

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -mod=mod -buildmode=plugin -o /plugin

FROM base AS coreBuild

WORKDIR /app

COPY go.mod go.sum /app/

COPY . /app/

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -mod=mod -o /microservice

FROM alpine
# Copies built plugin .so file into specified location from where it will be fetched by core
COPY --from=pluginBuild /plugin /etc/plugins

COPY --from=coreBuild /microservice /
COPY --from=coreBuild app/config.yaml .
COPY --from=coreBuild app/.env .

CMD ["./microservice"]
