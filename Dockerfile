FROM golang:1.10

WORKDIR /go/src/app
RUN bash -c 'curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh'
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
ADD pkg .
ADD cmd .
ADD version
RUN go install -v ./...

CMD ["app"]
