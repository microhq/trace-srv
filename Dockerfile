FROM alpine:3.2
ADD trace-srv /trace-srv
ENTRYPOINT [ "/trace-srv" ]
