FROM golang:1.16.4-alpine3.13 as builder

WORKDIR /go/src/github.com/htchan/WebHistory

COPY ./go.mod .
COPY ./go.sum .

RUN apk add gcc musl-dev libc-dev
RUN go mod download

COPY ./internal ./internal
COPY ./pkg ./pkg
COPY ./cmd/backend ./cmd/backend

RUN go build -v ./cmd/backend


FROM golang:1.16.3-alpine3.13

WORKDIR /usr/src/app

COPY --from=builder /go/src/github.com/htchan/WebHistory/backend .

ENV database_volume="/database/websites.db"

CMD ['./main']
