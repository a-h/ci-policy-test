FROM golang:1.14

COPY * /

WORKDIR / 

RUN go get && go build

ENTRYPOINT [ "/entrypoint.sh" ]