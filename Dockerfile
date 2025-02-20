FROM golang:1.24

WORKDIR /home

COPY src/server/go.mod src/server/go.sum ./
RUN go mod download

RUN ls -la

COPY src/server ./
COPY src/frontend ./frontend

RUN go build -v .

CMD ["./ascii"]
