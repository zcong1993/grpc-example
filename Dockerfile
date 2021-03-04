FROM alpine:latest

ADD ./bin /

CMD ["/server"]
