FROM golang:1.10 as build

RUN bash -c 'curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh'
#COPY Gopkg.toml Gopkg.lock ./
ADD . /go/src/github.com/flokkr/flokkr-operator
WORKDIR /opt
RUN curl https://storage.googleapis.com/kubernetes-helm/helm-v2.9.1-linux-amd64.tar.gz | tar xvz
WORKDIR /go/src/github.com/flokkr/flokkr-operator
RUN dep ensure -v --vendor-only
RUN go install -v ./...

FROM golang:1.10
COPY --from=build /opt/linux-amd64/helm /go/bin/helm
COPY --from=build /go/bin/flokkr-operator /go/bin/flokkr-operator
ADD launcher.sh /go/bin/launcher.sh
CMD ["/go/bin/launcher.sh"]
