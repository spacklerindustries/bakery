#FROM arm32v7/golang
FROM golang

COPY . /go/src/bakery
WORKDIR /go/src/bakery

EXPOSE 8080
RUN apt-get update && apt-get install kpartx -y

RUN go get -d -v ./...
RUN go build -o bakery *.go

ENV HTTP_PORT=8080
ENV GK_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZXhwIjowLCJ1c2VyaWQiOjEsInVzZXJuYW1lIjoiYWRtaW4ifQ.GQZFA7KICyo3-5xW4FOuwoNyJtjuGCQpIzzcPNgV-vM"
ENV BAKERY_ROOT=/app/bakery
ENV DB_PATH=/app/bakery/piInventory.db
ENV KPARTX_PATH=kpartx
ENV NFS_ADDRESS=127.0.0.1
ENV GK_SERVER="http://127.0.0.1:8080"
ENV TEMPLATE_PATH=/go/src/bakery/fileTemplates


VOLUME /app/bakery

CMD ["./bakery"]
