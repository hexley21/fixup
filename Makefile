MAKEFLAGS += --no-print-directory

SWAGGER_CONF=./configs/swagger-config.yaml

SRVCS=user catalog order


swag:
ifeq ($(OS),Windows_NT) 
	@$(MAKE) swag_batch
else
	@$(MAKE) swag_bash
endif


swag_batch:
	@echo urls: > $(SWAGGER_CONF)
	@for %%s in ($(SRVCS)) do ( \
		echo   - url: "./%%s.swagger.yaml" >> $(SWAGGER_CONF) && \
		echo     name: "%%s" >> $(SWAGGER_CONF) \
	)

swag_bash:
	@echo "urls:" > $(SWAGGER_CONF)
	@for svc in $(SRVCS); do \
		echo "  - url: "./$$svc.swagger.yaml"" >> $(SWAGGER_CONF); \
		echo "    name: "$$svc"" >> $(SWAGGER_CONF); \
	done
