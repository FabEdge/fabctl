FROM golang:1.17.13 as builder

ARG GOPROXY=https://goproxy.cn,direct

COPY . /fabctl

RUN cd /fabctl && GOPROXY=${GOPROXY} make fabctl

FROM alpine:3.15

COPY --from=builder /fabctl/_output/fabctl /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/fabctl"]