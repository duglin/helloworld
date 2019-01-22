FROM golang
COPY main.go /
RUN go build \
	-o /helloworld /main.go

FROM alpine
COPY --from=0 /helloworld /helloworld
ENTRYPOINT [ "/helloworld" ]
