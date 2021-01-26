FROM golang:1.14

COPY * /

WORKDIR / 

RUN go get && go build

CMD ["go", "run", "/main.go"]