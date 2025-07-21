# syntax=docker/dockerfile:1

FROM golang:1.24

WORKDIR /app

COPY . ./

ENV GOFLAGS=-mod=vendor
RUN go build -o money-transfers-api main/main.go

EXPOSE 8080

RUN ln -sf /usr/share/zoneinfo/Asia/Jerusalem /etc/localtime
RUN echo "Asia/Jerusalem" > /etc/timezone

CMD ./money-transfers-api
