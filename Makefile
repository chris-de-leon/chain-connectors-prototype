consumer.build:
	go build -o ./bin/apps/consumers/$(CONSUMER)/bin ./src/apps/consumers/$(CONSUMER)/main.go

producer.build:
	go build -o ./bin/apps/producers/$(PRODUCER)/bin ./src/apps/producers/$(PRODUCER)/main.go

consumer.run:
	go run ./src/apps/consumers/$(CONSUMER)/main.go

producer.run:
	go run ./src/apps/producers/$(PRODUCER)/main.go

docker.build.all:
	bash ./scripts/build.docker.sh "$(CONCURRENCY)"

docker.build.one:
	docker build --build-arg APP_DIR=$(APP_DIR) --tag $(TAG) .

docker.compose.up:
	docker compose up --build -d

docker.compose.down:
	docker compose down --remove-orphans

build.all:
	bash ./scripts/build.local.sh "$(CONCURRENCY)"

test.all:
	go test -v ./src/libs/...

protogen:
	bash ./scripts/protogen.sh

install:
	go get -v ./... && go mod tidy

upgrade:
	go get -v -u ./... && go mod tidy

clean:
	go clean -x -i -r -cache -modcache

