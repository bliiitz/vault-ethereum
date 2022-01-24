.ONESHELL:

DATE = $(shell date +'%s')

deploy-local-mac: export VAULT_ADDR=http://127.0.0.1:8200
deploy-local-mac: export SHA256=$(cat ./local/SHA256)

docker-build:
	docker build --build-arg always_upgrade="$(DATE)" -t cypherhat/vault-ethereum:latest .

run:
	docker-compose -f docker/docker-compose.yml up --build --remove-orphans

all: docker-build run