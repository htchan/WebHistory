FROM golang:1.16.4-alpine3.13 as builder

WORKDIR /go/src/github.com/htchan/WebHistory

COPY ./src ./src
COPY ./go.mod .

RUN apk add gcc musl-dev libc-dev
RUN go mod download

RUN go build -v ./src/main


FROM golang:1.16.3-alpine3.13

WORKDIR /usr/src/app

COPY --from=builder /go/src/github.com/htchan/WebHistory/main .

ENV database_volume="/database/websites.db"

CMD ['./main']
