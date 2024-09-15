ARG BASE_IMAGE=ubuntu:focal
FROM $BASE_IMAGE AS daiamteragent

RUN apt update && apt install iproute2 iputils-ping iperf3 wget git -y

RUN wget https://dl.google.com/go/go1.21.1.linux-amd64.tar.gz && tar -xvf go1.21.1.linux-amd64.tar.gz && mv go /usr/local
ENV GOROOT=/usr/local/go
RUN mkdir goproject
ENV GOPATH=/goproject
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

RUN git clone https://github.com/AbdallahRustom/DiameterTransfereAgent.git && cd DiameterTransfereAgent && \
git checkout AuthAccAsync

WORKDIR /DiameterTransfereAgent/

RUN cd /DiameterTransfereAgent \
&& go mod download \
&& go build -o myapp ./cmd/main.go

ENTRYPOINT ["./myapp"]