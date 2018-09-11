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

ENV KUBE_LATEST_VERSION="v1.11.2"

RUN curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl

RUN apt-get update && apt-get install -y jq python-yaml


ADD launcher.sh /go/bin/launcher.sh
ADD job/install.sh /go/bin/install.sh
ADD job/delete.sh /go/bin/delete.sh
RUN chmod o+w /go/bin
CMD ["/go/bin/launcher.sh"]
