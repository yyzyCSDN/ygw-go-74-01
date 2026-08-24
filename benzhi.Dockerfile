FROM golang:1.23.12 AS build
WORKDIR /src
ENV GOPROXY=off GOSUMDB=off
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o /cleanroom ./cmd/cleanroom

FROM golang:1.23.12
ENV GOPROXY=off GOSUMDB=off
ENV CGO_ENABLED=0
ENV GOFLAGS=-p=1
WORKDIR /app
COPY go.mod go.sum ./
COPY vendor ./vendor
COPY . .
COPY --from=build /cleanroom /usr/local/bin/cleanroom
EXPOSE 8090
CMD ["/usr/local/bin/cleanroom", "-addr", "0.0.0.0:8090", "-dir", "/app/data", "-web-dir", "/app/web"]
