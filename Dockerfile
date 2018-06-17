#FROM arm32v7/golang
FROM golang

COPY . /go/src/bakery
WORKDIR /go/src/bakery

EXPOSE 8080
RUN apt-get update && apt-get install kpartx -y

RUN go get -d -v ./...
RUN go build -o bakery *.go

ENV HTTP_PORT=8080
ENV MQTT_SERVER=127.0.0.1:1883
ENV NFS_ADDRESS=127.0.0.1
ENV BAKERY_ROOT=/app/bakery
ENV DB_PATH=/app/bakery/piDb.db
ENV KPARTX_PATH=kpartx
ENV TEMPLATE_PATH=/go/src/bakery/fileTemplates

VOLUME /app/bakery

CMD ["./bakery"]
