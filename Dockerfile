FROM alpine:3

ARG GIT_SHA

COPY ./scepserver-linux-amd64 /usr/bin/scepserver

EXPOSE 9001

VOLUME ["/depot"]

ENTRYPOINT ["scepserver"]
