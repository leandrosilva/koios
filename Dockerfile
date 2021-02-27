FROM golang:1.14

WORKDIR /go/src/app
COPY . .

# RUN go get -d -v ./...
# RUN go install -v ./...

RUN go build -o koios ./.
RUN ls -la .

EXPOSE 6660

CMD ["./koios"]
