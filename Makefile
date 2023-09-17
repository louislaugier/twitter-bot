.DEFAULT_GOAL := dev

.PHONY: dev
dev:
	docker-compose up -d

.PHONY: stop
stop:
	docker-compose stop $(s) && docker-compose rm -f $(s)

.PHONY: build
build:
	docker-compose build --no-cache --build-arg DOCKERFILE=Dockerfile.dev $(s) && docker-compose up -d $(s) --force-recreate

.PHONY: rebuild
rebuild:
	make stop s=$(s)
	make build s=$(s)

################################@

.PHONY: freelancechain
freelancechain:
	make rebuild s=freelancechain

# .PHONY: webstudio
# webstudio:
# 	make rebuild s=webstudio