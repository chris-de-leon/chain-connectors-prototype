build.consumer:
	go build -o ./bin/apps/consumers/$(CONSUMER)/bin ./src/apps/consumers/$(CONSUMER)/main.go

build.producer:
	go build -o ./bin/apps/producers/$(PRODUCER)/bin ./src/apps/producers/$(PRODUCER)/main.go

build.all:
	bash ./scripts/build.local.sh "$(CONCURRENCY)"

docker.build.all:
	bash ./scripts/build.docker.sh "$(CONCURRENCY)"

docker.build.one:
	docker build --build-arg APP_DIR=$(APP_DIR) --tag $(TAG) .

protogen:
	bash ./scripts/protogen.sh

install:
	go get -v ./... && go mod tidy

upgrade:
	go get -v -u ./... && go mod tidy

clean:
	go clean -x -i -r -cache -modcache

