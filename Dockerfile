FROM golang:latest as builder
WORKDIR /go/src/github.com/cubeee/site
RUN go-wrapper download github.com/jimlawless/cfg \
 && go-wrapper download github.com/flosch/pongo2 \
 && go-wrapper download goji.io \
 && go-wrapper download goji.io/pat \
 && go-wrapper download github.com/fsnotify/fsnotify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o build/site .

FROM alpine:latest
RUN apk --no-cache add ca-certificates bash
WORKDIR /root/
COPY --from=builder /go/src/github.com/cubeee/site/build/site ./
COPY --from=builder /go/src/github.com/cubeee/site/resources ./resources/
RUN chmod +x /root/site;
CMD ["/root/site"]