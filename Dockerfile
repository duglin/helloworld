FROM golang
ARG gitsha
COPY main.go /
RUN GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 go build \
    -ldflags "-w -extldflags -static" \
    -tags netgo -installsuffix netgo \
    -o /helloworld /main.go

FROM alpine
COPY --from=0 /helloworld /helloworld
ENTRYPOINT [ "/helloworld" ]
