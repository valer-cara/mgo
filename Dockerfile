# Stage 1: Build
FROM golang AS build

ENV ROOT "/go/src/github.com/valer-cara/mgo"

WORKDIR $ROOT
ADD Gopkg.lock Gopkg.toml ./

RUN go get -u github.com/golang/dep/cmd/dep \
 && dep ensure --vendor-only

ADD . .
RUN make install


# Stage 2: Actual image
FROM ubuntu

# some stuff required by helm install plugin
RUN apt-get update && apt-get install -y ca-certificates git curl make

## TODO: match the version of helm to cluster, needs to be handled on a higher level
RUN curl --output - https://storage.googleapis.com/kubernetes-helm/helm-v2.13.1-linux-amd64.tar.gz | tar zxvf - \
 && mv linux-amd64/helm /usr/local/bin/helm

ENV HELM_HOME /root/.helm
RUN helm init --client-only

COPY ./script/start.sh /start.sh
COPY --from=build /go/bin/mgo /mgo

CMD ["/mgo"]
