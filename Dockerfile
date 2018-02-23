FROM golang:latest as builder
RUN go get -d -v github.com/ymatsiuk/swarm-registrator
WORKDIR /go/src/github.com/ymatsiuk/swarm-registrator
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o swarm-registrator .

FROM gruebel/upx:latest as upx
COPY --from=builder /go/src/github.com/ymatsiuk/swarm-registrator/swarm-registrator /swarm-registrator.orig
RUN upx --best --lzma -o /swarm-registrator /swarm-registrator.orig

FROM scratch
COPY --from=upx /swarm-registrator /swarm-registrator
CMD ["/swarm-registrator"]