FROM arm32v7/golang as builder
#FROM golang

COPY . /go/src/bakery
WORKDIR /go/src/bakery

RUN go get -d -v ./...
RUN go build -o bakery *.go

FROM arm32v7/ubuntu

RUN mkdir -p /go/src/bakery/fileTemplates
WORKDIR /go/src/bakery
COPY --from=builder /go/src/bakery/bakery .
COPY --from=builder /go/src/bakery/fileTemplates/. fileTemplates/.

EXPOSE 8080
RUN apt-get update && apt-get install kpartx nfs-kernel-server nfs-common -y

ENV HTTP_PORT=8080
ENV BAKERY_ROOT=/app/bakery
ENV DB_PATH=/app/bakery/piInventory.db
ENV KPARTX_PATH=kpartx
ENV NFS_ADDRESS=127.0.0.1
ENV BUSHWOOD_SERVER="http://127.0.0.1:8080"
ENV BUSHWOOD_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZXhwIjowLCJ1c2VyaWQiOjEsInVzZXJuYW1lIjoiYWRtaW4ifQ.GQZFA7KICyo3-5xW4FOuwoNyJtjuGCQpIzzcPNgV-vM"
ENV TEMPLATE_PATH=/go/src/bakery/fileTemplates


VOLUME /app/bakery

CMD ["./bakery"]
