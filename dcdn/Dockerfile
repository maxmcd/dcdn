FROM yukinying/chrome-headless-browser:latest

RUN apt-get update && apt-get install -y --no-install-recommends \
        g++ \
        gcc \
        libc6-dev \
        make \
        git \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.8.3

# https://github.com/docker-library/golang/blob/64b88dc3e9d83e71eafc000fed1f0d5e289b3e65/1.8/Dockerfile
RUN set -eux; \
    \
    dpkgArch="$(dpkg --print-architecture)"; \
    goRelArch=linux-amd64; \
    goRelSha256=1862f4c3d3907e59b04a757cfda0ea7aa9ef39274af99a784f5be843c80c6772; \
    url="https://golang.org/dl/go${GOLANG_VERSION}.${goRelArch}.tar.gz"; \
    wget -O go.tgz "$url"; \
    echo "${goRelSha256} *go.tgz" | sha256sum -c -; \
    tar -C /usr/local -xzf go.tgz; \
    rm go.tgz; \
    \
    if [ "$goRelArch" = 'src' ]; then \
        echo >&2; \
        echo >&2 'error: UNIMPLEMENTED'; \
        echo >&2 'TODO install golang-any from jessie-backports for GOROOT_BOOTSTRAP (and uninstall after build)'; \
        echo >&2; \
        exit 1; \
    fi; \
    \
    export PATH="/usr/local/go/bin:$PATH"; \
    go version

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

RUN go get -u github.com/golang/dep/cmd/dep

RUN mkdir -p /go/src/github.com/maxmcd/dcdn/dcdn/
WORKDIR /go/src/github.com/maxmcd/dcdn/dcdn/
VOLUME /go/src/github.com/maxmcd/dcdn/dcdn/

COPY . /go/src/github.com/maxmcd/dcdn/dcdn/

RUN dep ensure

ENTRYPOINT ["sleep", "1000"]

