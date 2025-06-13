FROM golang:1.24.4

LABEL maintainer="Roberto Manfreda <robertomanfreda@protonmail.com>"

WORKDIR /app

COPY . .

RUN go build -o main .

CMD ["./main"]