FROM alpine:latest
WORKDIR /usr/src/app

RUN apk add postgresql-client python3

COPY entrypoint.sh cleanup.py .

CMD '/usr/src/app/entrypoint.sh'