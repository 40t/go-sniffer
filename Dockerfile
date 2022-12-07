###############################################################################
#  Docker image for build go-sniffer library
###############################################################################

FROM golang:1.16.15-stretch

RUN apt update
RUN apt -y install libpcap-dev
RUN mkdir -p /usr/local/go/src/go-sniffer
WORKDIR /usr/local/go/src/go-sniffer
COPY core core
COPY pkg pkg
COPY plugSrc plugSrc
COPY plug plug
COPY main.go main.go
COPY go.mod go.mod
COPY go.sum go.sum
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o go-sniffer
ENTRYPOINT  [ "/usr/local/go/src/go-sniffer/go-sniffer" ]


################################################################################
##  Docker image for build go-sniffer
################################################################################
#FROM golang:1.16.15-stretch
#
#RUN apt update;\
#    apt install -y libpcap-dev
#
#RUN mkdir /app
#WORKDIR /app
#COPY --from=tool /usr/local/go/src/go-sniffer/go-sniffer /app/go-sniffer
#ENTRYPOINT  [ "/app/go-sniffer" ]