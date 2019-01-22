FROM golang
ARG gitsha
COPY main.go /
# RUN GO_EXTLINK_ENABLED=0 CGO_ENABLED=0 go build \
#     -ldflags "-w -extldflags -static" \
#     -tags netgo -installsuffix netgo \
#     -o /helloworld /main.go
RUN go build -o /helloworld /main.go

FROM golang
COPY --from=0 /helloworld /helloworld
ENTRYPOINT [ "/helloworld" ]
