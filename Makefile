include .envrc

MAKEFLAGS += --no-print-directory

SWAGGER_CONF=./configs/swagger-config.yaml

SRVCS=user catalog order chat


build:
ifeq ($(OS),Windows_NT) 
	@$(MAKE) build/batch
else
	@$(MAKE) build/bash
endif


swag:
ifeq ($(OS),Windows_NT) 
	@$(MAKE) swag/batch
else
	@$(MAKE) swag/bash
endif


compose: build
	@docker-compose up --build --remove-orphans


sqlc:
	@sqlc generate -f ./sql/$(db)/sqlc.yml


migrate-up:
	@$(MAKE) migrate db=$(db) way=up

migrate-down:
	@$(MAKE) migrate db=$(db) way=down

migrate:
	migrate -path ./sql/$(db)/migrations -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT_$(db)}/$(db)?sslmode=disable" -verbose $(way)

migrate-init:
	migrate create -ext sql -dir ./$(db)-service/sql/migrations/ -seq init_schema


# PLATFORM SPECIFIC

build/batch:
	@for %%s in ($(SRVCS)) do ( \
		set CGO_ENABLED=0&& set GOOS=linux&& go build -o ./bin/%%s ./cmd/%%s/main.go \
	)

build/bash:
	@for svc in $(SRVCS); do \
		CGO_ENABLED=0 GOOS=linux go build -o ./bin/$$svc ./cmd/$$svc/main.go; \
	done


swag/batch:
	@echo urls: > $(SWAGGER_CONF)
	@for %%s in ($(SRVCS)) do ( \
		echo   - url: "./%%s.swagger.yaml" >> $(SWAGGER_CONF) && \
		echo     name: "%%s" >> $(SWAGGER_CONF) \
	)

swag/bash:
	@echo "urls:" > $(SWAGGER_CONF)
	@for svc in $(SRVCS); do \
		echo "  - url: "./$$svc.swagger.yaml"" >> $(SWAGGER_CONF); \
		echo "    name: "$$svc"" >> $(SWAGGER_CONF); \
	done
