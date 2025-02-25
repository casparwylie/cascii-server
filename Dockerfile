FROM golang:1.24 AS server

WORKDIR /home

COPY src/server/go.mod src/server/go.sum ./
RUN go mod download

RUN ls -la

COPY src/server ./
COPY src/frontend ./frontend

RUN go build -v .
RUN rm *.go

CMD ["./ascii"]

FROM server AS tests

COPY src/tests ./
COPY db/migrations ./migrations

RUN apt-get update \
    && apt-get install -y default-mysql-client

RUN go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

ENTRYPOINT ["./setup.sh"]
