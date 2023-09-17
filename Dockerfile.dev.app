FROM golang:1.18-alpine

ENV env=dev

WORKDIR /go/src/twitter-bot

COPY . .

RUN go mod download && go install github.com/cosmtrek/air@latest

CMD ["air", "."]