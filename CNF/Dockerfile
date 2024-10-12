FROM dockerproxy.repos.tech.orange/alpine:3.19.1

# Install necessary packages
RUN apk add --no-cache iproute2=5.10.0-r0 iputils=1.9-r0 git=2.34.1-r0 wget=1.21.1-r0

# Set up Go environment
RUN wget https://dl.google.com/go/go1.21.1.linux-amd64.tar.gz && tar -xvf go1.21.1.linux-amd64.tar.gz && mv go /usr/local
ENV GOROOT=/usr/local/go
RUN mkdir goproject
ENV GOPATH=/goproject
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Clone the repository and checkout the desired branch
COPY ./ /DiameterTransfereAgent/

WORKDIR /DiameterTransfereAgent/

# Download dependencies and build the application
RUN go mod download && \
    go build -o myapp ./cmd/main.go

# Set the entry point
ENTRYPOINT ["./myapp"]