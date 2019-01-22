all:
	docker build -t duglin/helloworld .
	docker push duglin/helloworld
