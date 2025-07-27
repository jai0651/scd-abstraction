FROM golang:1.24

WORKDIR /app

COPY . .

RUN go mod download

CMD ["go", "test", "-bench=.", "./benchmark"] 