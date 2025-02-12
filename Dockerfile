FROM golang:1.20
WORKDIR /app
COPY . .
RUN go build -o postcomm .
EXPOSE 8080
CMD ["./postcomm"]