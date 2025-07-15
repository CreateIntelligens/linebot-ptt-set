FROM golang:1.21-alpine

WORKDIR $GOPATH/src/mong0520/linebot-ptt-set
COPY . $GOPATH/src/mong0520/linebot-ptt-set
RUN GO111MODULE=on go build

EXPOSE 5050
ENTRYPOINT ["./linebot-ptt-set"]
CMD ["./linebot-ptt-set"]
