.PHONY: static

docker_compose_path = "docker-compose.yml"

DC = docker-compose -f $(docker_compose_path)

format:
	gofumpt -l -w .
	gci write . --skip-generated -s standard -s default

lint:
	golangci-lint run

app-up:
	$(DC) up -d --build

down:
	$(DC) down