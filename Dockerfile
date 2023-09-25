FROM golang:1.20

WORKDIR /app
COPY . .
RUN make

CMD ["./main"]
