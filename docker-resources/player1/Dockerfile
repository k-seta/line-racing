FROM golang:1.22.1

RUN apt update && apt install -y curl

ENV GO111MODULE=on
ENV ROOT=/go/src/app
WORKDIR ${ROOT}

RUN go install github.com/cosmtrek/air@latest

EXPOSE 8080

CMD ["air"]
