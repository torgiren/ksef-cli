run:
	@echo "Running ksef-cli..."
	go run ksef-cli.go

build: ksef-cli.go
	@echo "Building ksef-cli..."
	go build -o ksef-cli ksef-cli.go

update-openapi:
	@echo "Downloading latest OpenAPI specification..."
	curl -sS -o internal/ksefapi/openapi.json https://api.ksef.mf.gov.pl/docs/v2/openapi.json


.PHONY: update-openapi build run
