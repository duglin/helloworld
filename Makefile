all: .helloworld .rebuild load

APP_IMAGE     ?= duglin/helloworld
REBUILD_IMAGE ?= duglin/rebuild

load: load.go
	go build -o load load.go

.rebuild: rebuild.go Dockerfile.rebuild
	go build -o /dev/null rebuild.go     # Fail fast for compilation errors
	docker build -f Dockerfile.rebuild -t $(REBUILD_IMAGE) .
	docker push $(REBUILD_IMAGE)
	touch .rebuild                       # Just to keep `make` happy

.helloworld: helloworld.go Dockerfile
	go build -o /dev/null helloworld.go  # Fail fast for compilation errors
	docker build -t $(APP_IMAGE) .       # Do the real build and create image
	docker push $(APP_IMAGE)
	touch .helloworld                    # Just to keep `make` happy
