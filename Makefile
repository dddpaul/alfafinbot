.PHONY: all build release

IMAGE=dddpaul/alfafinbot

all: build

build-alpine:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./bin/bot ./main.go

build:
	@docker build --tag=${IMAGE} .

debug:
	@docker run -it --entrypoint=sh ${IMAGE}

release: build
	@echo "Tag image with version $(version)"
	@docker tag ${IMAGE} ${IMAGE}:$(version)

push: release
	@docker push ${IMAGE}
	@docker push ${IMAGE}:$(version)
