include .env
export PROJECT_ROOT=$(shell pwd)

ENV_FILE := .env
DC := docker compose --env-file $(ENV_FILE)
export

env-up:
	@$(DC) up -d $(POSTGRES_CONTAINER)
	@$(DC) up -d $(NATS_CONTAINER)
env-down:
	@$(DC) stop $(POSTGRES_CONTAINER)	
	@$(DC) stop $(NATS_CONTAINER)
env-cleanup:
	@$(DC) down -v
	@rm -rf ${PROJECT_ROOT}$(POSTGRES_FOLDER) && \
	@rm -rf ${PROJECT_ROOT}$(NATS_FOLDER)
	@echo "All data cleaned up for $(SERVICE_NAME)."

env-port-forward:
	@$(DC) up -d $(PORT_FORWARDER)

env-port-close:
	@$(DC) stop $(PORT_FORWARDER)
	@$(DC) rm -f $(PORT_FORWARDER)

migrate-create:
ifndef seq
	$(error Не указано имя миграции! Использование: make migrate-create seq=add_users_table)
endif
	@$(DC) run --rm $(MIGRATION_CONTAINER) \
		create \
		-ext sql \
		-dir $(MIGRATION_FOLDER) \
		-seq "$(seq)"

migrate-up:
	@$(MAKE) migrate-action action="$(if $(action),$(action),up)"

migrate-down:
	@$(MAKE) migrate-action action="$(if $(action),$(action),down 1)"

migrate-drop:
	@$(MAKE) migrate-action action="drop -f"

migrate-version:
	@$(MAKE) migrate-action action=version

migrate-action:
	@$(DC) run --rm $(MIGRATION_CONTAINER) \
		-path $(MIGRATION_FOLDER) \
		-database "postgresql://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_CONTAINER):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
		$(action)
run:
	@go run ${PROJECT_ROOT}/cmd/${SERVICE_NAME}/main.go 

deploy:
	@docker compose up -d --build notification
	@docker image prune -f
deploy-down:
	@docker compose down
	@docker image prune -f

# docker run --rm -it --network host natsio/nats-box nats sub ">"
