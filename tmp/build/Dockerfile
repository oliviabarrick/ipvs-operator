FROM alpine:3.6

RUN adduser -D ipvs-operator
USER ipvs-operator

ADD tmp/_output/bin/ipvs-operator /usr/local/bin/ipvs-operator
