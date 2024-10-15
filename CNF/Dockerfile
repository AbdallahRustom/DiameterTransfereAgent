FROM dockerproxy.repos.tech.orange/alpine:3.20.3 AS builder
ARG SRC_REPO_USER
ARG SRC_REPO_PWD
ARG SRC_REPO_URL=gitlab.tech.orange/win/oinis/olb-2g6/Diameter-Translation-Agent.git
ARG SRC_REPO_BRANCH

# Install necessary packages
RUN apk add --no-cache iproute2=6.9.0-r0 \ 
    iputils=20240117-r0 \ 
    git=2.45.2-r0 \
    wget=1.24.5-r0 \
    libssl3=3.3.2-r0 \
    libcrypto3=3.3.2-r0
    
# Set up Go environment
RUN wget --max-redirect=0 --progress=dot:giga https://dl.google.com/go/go1.21.2.linux-amd64.tar.gz && \
    tar -xvf go1.21.2.linux-amd64.tar.gz && \
    mv go /usr/local
ENV GOROOT=/usr/local/go
RUN mkdir goproject
ENV GOPATH=/goproject
ENV PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# Clone the repository and checkout the desired branch
RUN git clone https://${SRC_REPO_USER}:${SRC_REPO_PWD}@${SRC_REPO_URL} DiameterTransfereAgent;

WORKDIR /DiameterTransfereAgent/
RUN git checkout ${SRC_REPO_BRANCH}

# Download dependencies and build the application
RUN go mod download && \
    go build -o myapp ./cmd/main.go


FROM dockerproxy.repos.tech.orange/alpine:3.20.3
RUN apk add --no-cache iproute2=6.9.0-r0 \ 
iputils=20240117-r0 \ 
git=2.45.2-r0 \
wget=1.24.5-r0 \
libssl3=3.3.2-r0 \
libcrypto3=3.3.2-r0

RUN addgroup -g 1000 dta \
&& adduser -D -G dta dta -u 1000 -s /bin/sh \
&& mkdir /dta && chown dta:dta /dta
USER 1000
WORKDIR /dta

COPY --chown=dta:dta --from=builder /DiameterTransfereAgent/myapp ./myapp
COPY --chown=dta:dta --from=builder /DiameterTransfereAgent/config.json ./config.json
COPY --chown=dta:dta --from=builder /DiameterTransfereAgent/dictionary ./dictionary

ENTRYPOINT ["./myapp"]