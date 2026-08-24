FROM golang:1.26.2

ENV GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
WORKDIR /workspace
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
RUN go build -mod=vendor -o /usr/local/bin/kilnguard ./cmd/kilnguard
CMD ["/usr/local/bin/kilnguard"]
