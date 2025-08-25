.PHONY: all build schema

all: build schema

build:
	go build -o gencmd .

schema:
	go run ./hack/jsonschema/main.go > config-schema.json
