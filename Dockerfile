#FROM arm32v7/golang
FROM golang

COPY . /go/src/bakery
WORKDIR /go/src/bakery

EXPOSE 8080
RUN apt-get update && apt-get install kpartx -y

RUN go get -d -v ./...
RUN go build -o main *.go

ENV NFS_ADDRESS=127.0.0.1
ENV BAKERY_ROOT=/app/bakery
ENV DB_PATH=/app/bakery/piDb.db
ENV PPI_PATH=/app/bakery/ppi
ENV PPI_CONFIG_PATH=/app/bakery/ppiConfig.json
ENV KPARTX_PATH=kpartx

VOLUME /app/bakery

CMD ["./main"]

