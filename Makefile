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

swag-config:
ifeq ($(OS),Windows_NT) 
	@$(MAKE) swag-config/batch
else
	@$(MAKE) swag/bash
endif

swag-gen:
ifeq ($(OS),Windows_NT) 
	@$(MAKE) swag-gen/batch svc=$(svc)
else
	@$(MAKE) swag-gen/bash svc=$(svc)
endif

test: 
	go test -cover ./internal/...
	@echo "REPOSITORY TESTS:"
	go test -cover ./internal/user/repository/ -mp="${CURDIR}/sql/user/migrations"
	go test -cover ./internal/catalog/repository -mp="${CURDIR}/sql/catalog/migrations"

compose: build
	@docker compose up --build --remove-orphans

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

swag-gen/bash:
	@swag init --dir cmd/$(svc)/,internal/$(svc)/delivery/http,pkg/http/rest,internal/common --parseDependency --output ./api/swagger --outputTypes yaml
	mv ./api/swagger/swagger.yaml ./api/swagger/$(svc).swagger.yaml

swag-gen/batch:
	@swag init --dir cmd/$(svc)/,internal/$(svc)/delivery/http,pkg/http/rest,internal/common --parseDependency --output ./api/swagger --outputTypes yaml
	@if exist api\swagger\$(svc).swagger.yaml del api\swagger\$(svc).swagger.yaml
	ren api\swagger\swagger.yaml $(svc).swagger.yaml

swag-config/bash:
	@echo "urls:" > $(SWAGGER_CONF)
	@for svc in $(SRVCS); do \
		echo "  - url: "./$$svc.swagger.yaml"" >> $(SWAGGER_CONF); \
		echo "    name: "$$svc"" >> $(SWAGGER_CONF); \
	done

swag-config/batch:
	@echo urls: > $(SWAGGER_CONF)
	@for %%s in ($(SRVCS)) do ( \
		echo   - url: "./%%s.swagger.yaml" >> $(SWAGGER_CONF) && \
		echo     name: "%%s" >> $(SWAGGER_CONF) \
	)
