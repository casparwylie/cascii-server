FROM golang:1.24 AS server

WORKDIR /home

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/server ./
COPY src/frontend/*.js ./frontend/
COPY src/frontend/cascii-core/cascii.html ./frontend/cascii-core/cascii.html

RUN go build -v .

CMD ["./cascii-server"]

FROM server AS tests

COPY src/tests ./
COPY db/migrations ./migrations

RUN apt-get update \
    && apt-get install -y default-mysql-client

RUN go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

ENTRYPOINT ["./setup.sh"]
