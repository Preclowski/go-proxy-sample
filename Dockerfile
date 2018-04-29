FROM golang:latest AS build-env

WORKDIR /go/src/app/app/app
ADD . .

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep ensure
RUN go build -o /go/bin/app
RUN chmod +x /go/bin/app

FROM scratch

COPY --from=build-env /go/bin/app /go/bin/app

ENTRYPOINT ["/go/bin/app"]