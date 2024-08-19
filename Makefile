include .envrc

MAKEFLAGS += --no-print-directory

SWAGGER_CONF=./configs/swagger-config.yaml

SRVCS=user catalog order


migrate-up:
	@$(MAKE) migrate db=$(db) way=up

migrate-down:
	@$(MAKE) migrate db=$(db) way=down

migrate:
	migrate -path ./$(db)-service/sql/migrations -database "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@localhost:${POSTGRES_PORT_$(db)}/$(db)?sslmode=disable" -verbose $(way)


migrate-init:
	migrate create -ext sql -dir ./$(db)-service/sql/migrations/ -seq init_schema


sqlc:
	@sqlc generate -f ./$(db)-service/sqlc.yml


compose:
	@docker-compose up --build --remove-orphans


swag:
ifeq ($(OS),Windows_NT) 
	@$(MAKE) swag/batch
else
	@$(MAKE) swag/bash
endif

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
