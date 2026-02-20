COMPOSE_FILE := docker/compose.yaml
PROFILE ?= dev
SIM_MODE ?= bot-orders
# Supported profiles: dev, loadtest, prodlike

.PHONY: up down restart logs ps build pull config

up:
	SIM_MODE=$(SIM_MODE) docker compose -f $(COMPOSE_FILE) --profile $(PROFILE) up --build

down:
	docker compose -f $(COMPOSE_FILE) --profile $(PROFILE) down

restart: down up

logs:
	docker compose -f $(COMPOSE_FILE) --profile $(PROFILE) logs -f

ps:
	docker compose -f $(COMPOSE_FILE) --profile $(PROFILE) ps

build:
	SIM_MODE=$(SIM_MODE) docker compose -f $(COMPOSE_FILE) --profile $(PROFILE) build

pull:
	docker compose -f $(COMPOSE_FILE) --profile $(PROFILE) pull

config:
	SIM_MODE=$(SIM_MODE) docker compose -f $(COMPOSE_FILE) --profile $(PROFILE) config
