FROM ubuntu:24.04 As builder

RUN apt update -y
RUN apt install wget tar gzip git -y

RUN apt install build-essential libcap-dev pkg-config libsystemd-dev -y
RUN wget -P /tmp https://github.com/ioi/isolate/archive/master.tar.gz && tar -xzvf /tmp/master.tar.gz -C / > /dev/null
RUN make -C /isolate-master isolate
ENV PATH="/isolate-master:$PATH"

COPY ./configs/isolate_default.cf /usr/local/etc/isolate
# creates sandbox 0
RUN isolate --init

# go installation
WORKDIR /usr/local
ARG GO_VERSION=1.22.0
ARG GO_ARCH=linux-amd64
RUN wget -q -o /dev/null https://go.dev/dl/go${GO_VERSION}.${GO_ARCH}.tar.gz
RUN tar -xzf go${GO_VERSION}.${GO_ARCH}.tar.gz

ENV GOROOT=/usr/local/go
ENV GOPATH=/root/go
ENV PATH=${GOROOT}/bin:${GOPATH}/bin:${PATH}

WORKDIR /app
ARG SERVICE_PORT
ENV SERVICE_PORT ${SERVICE_PORT}

COPY ./src .

RUN go mod download
RUN go build -o ./cmd/server *.go

EXPOSE ${SERVICE_PORT}

CMD ["./cmd/server"]
