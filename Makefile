all: .helloworld .rebuild load

load: load.go
	go build -o load load.go

.rebuild: rebuild.go Dockerfile.rebuild rebuild.sh
	go build -o /dev/null rebuild.go     # Fail fast for compilation errors
	docker build -f Dockerfile.rebuild -t duglin/rebuild .
	docker push duglin/rebuild
	touch .rebuild                       # Just to keep `make` happy

.helloworld: helloworld.go Dockerfile
	go build -o /dev/null helloworld.go  # Fail fast for compilation errors
	docker build -t duglin/helloworld .  # Do the real build and create image
	docker push duglin/helloworld
	touch .helloworld                    # Just to keep `make` happy
