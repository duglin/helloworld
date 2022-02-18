FROM golang
COPY helloworld.go /
RUN go build -o /helloworld /helloworld.go
FROM duglin/ubuntu
COPY --from=0 /helloworld /helloworld
ENTRYPOINT [ "/helloworld" ]
