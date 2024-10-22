include .envrc

MAKEFLAGS += --no-print-directory

SWAGGER_CONF=./configs/swagger-config.yaml

SVCS=user catalog order chat

# Generates swagger config file according to SVCS
swag-config:
ifeq ($(OS),Windows_NT) 
	@$(MAKE) swag-config/batch
else
	@$(MAKE) swag/bash
endif

# Generates swagger documentation file according to $(svc)
swag-gen:
ifeq ($(OS),Windows_NT) 
	@$(MAKE) swag-gen/batch svc=$(svc)
else
	@$(MAKE) swag-gen/bash svc=$(svc)
endif

# Run all tests
test: 
	go test -cover ./internal/...
	@echo "REPOSITORY TESTS:"
	$(MAKE) test-repo

# Only test repositories
test-repo:
	go test -cover ./internal/user/repository/ -mp="${CURDIR}/sql/user/migrations"
	go test -cover ./internal/catalog/repository -mp="${CURDIR}/sql/catalog/migrations"

# Genrates sqlc files according to $(db)
sqlc:
	@sqlc generate -f ./sql/$(db)/sqlc.yml

# Migrates database up according to $(db) and credentials from .envrc
migrate-up:
	@$(MAKE) migrate db=$(db) way=up

# Migrates database down according to $(db) and credentials from .envrc
migrate-down:
	@$(MAKE) migrate db=$(db) way=down

# Migrates database according to $(way){up/down} and $(db), gets credentials and values form .envrc
migrate:
	migrate -path ./sql/$(db)/migrations -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT_$(db)}/$(db)?sslmode=disable" -verbose $(way)

# Initializes migrate up & down files for $(db)
migrate-init:
	migrate create -ext sql -dir ./$(db)-service/sql/migrations/ -seq init_schema


# PLATFORM SPECIFIC
swag-gen/bash:
	@swag init --dir cmd/$(svc)/,internal/$(svc)/delivery/http,pkg/http/rest,internal/common --parseDependency --output ./api/swagger --outputTypes yaml
	mv ./api/swagger/swagger.yaml ./api/swagger/$(svc).swagger.yaml

swag-gen/batch:
	@swag init --dir cmd/$(svc)/,internal/$(svc)/delivery/http,pkg/http/rest,internal/common --parseDependency --output ./api/swagger --outputTypes yaml
	@if exist api\swagger\$(svc).swagger.yaml del api\swagger\$(svc).swagger.yaml
	ren api\swagger\swagger.yaml $(svc).swagger.yaml

swag-config/bash:
	@echo "urls:" > $(SWAGGER_CONF)
	@for svc in $(SVCS); do \
		echo "  - url: "./$$svc.swagger.yaml"" >> $(SWAGGER_CONF); \
		echo "    name: "$$svc"" >> $(SWAGGER_CONF); \
	done

swag-config/batch:
	@echo urls: > $(SWAGGER_CONF)
	@for %%s in ($(SVCS)) do ( \
		echo   - url: "./%%s.swagger.yaml" >> $(SWAGGER_CONF) && \
		echo     name: "%%s" >> $(SWAGGER_CONF) \
	)
