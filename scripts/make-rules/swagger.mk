
.PHONY: swagger.run
swagger.run: tools.verify.swagger
	@echo "===========> Generating swagger API docs"
	#@swagger generate spec --scan-models -w $(PROJECT_ROOT)/cmd/gen-swagger-type-docs -o $(PROJECT_ROOT)/api/swagger/kubernetes.yaml
	@swagger mixin `find $(PROJECT_ROOT)/api/openapi -name "*.swagger.json"` \
		-q                                                    \
		--keep-spec-order                                     \
		--format=yaml                                         \
		--ignore-conflicts                                    \
		-o $(PROJECT_ROOT)/api/swagger/swagger.yaml
	@echo "Generated at: $(PROJECT_ROOT)/api/swagger/swagger.yaml"

.PHONY: swagger.serve
swagger.serve: tools.verify.swagger
	@swagger serve -F=redoc --no-open --port 65534 $(PROJECT_ROOT)/api/swagger/swagger.yaml