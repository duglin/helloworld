FROM golang
ARG gitsha
COPY helloworld.go /
RUN go build -o /helloworld /helloworld.go
ENTRYPOINT [ "/helloworld" ]
