FROM golang:1.22.1

RUN apt update && apt install -y curl
EXPOSE 8080

COPY ./line-racing-bot-dev main

CMD ["./main"]
