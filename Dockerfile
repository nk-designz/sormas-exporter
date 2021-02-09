FROM golang:latest as builder
WORKDIR /go/src/sormas-exporter
COPY . .
RUN go get -v . 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o /go/bin/sormas_exporter

FROM alpine 
COPY --from=builder /go/bin/sormas_exporter /sormas_exporter
ENV HOST postgres
ENV PORT 5432
ENV USER sormas_user
ENV PASSWORD password
ENV RETRY 5 
CMD /sormas_exporter -host=${HOST} -port=${PORT} -user=${USER} -password=${PASSWORD} -retry=${RETRY}