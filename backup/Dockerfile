FROM alpine:latest
WORKDIR /usr/src/app

RUN apk add postgresql-client

COPY entrypoint.sh .

CMD '/usr/src/app/entrypoint.sh'