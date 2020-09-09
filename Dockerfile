FROM golang:1.15
ENV GOBIN=$GOPATH/bin

# Downloading gcloud package
RUN curl https://dl.google.com/dl/cloudsdk/release/google-cloud-sdk.tar.gz > /tmp/google-cloud-sdk.tar.gz

# Installing the package
RUN mkdir -p /usr/local/gcloud \
  && tar -C /usr/local/gcloud -xvf /tmp/google-cloud-sdk.tar.gz \
  && /usr/local/gcloud/google-cloud-sdk/install.sh

# Adding the package path to local
ENV PATH $PATH:/usr/local/gcloud/google-cloud-sdk/bin

# kubectl
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.16.0/bin/linux/amd64/kubectl
RUN chmod +x ./kubectl
RUN mv ./kubectl /usr/local/bin/kubectl
RUN kubectl version --client


ENV KUBECONFIG ./config

#install rio
RUN curl -sfL https://get.rio.io | sh -
RUN rio install
RUN rio -n rio-system pods

WORKDIR /go/src/app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
#COPY go.sum .

RUN go mod download

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...


RUN ls $GOPATH
RUN echo $GOPATH
RUN ls $GOBIN

CMD go run main.go
#CMD $GOBIN/fundamenv
