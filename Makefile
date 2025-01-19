setup:
	@make keys

# Sets up the development environment by downloading dependencies installing
# pre-commit hooks, generating keys, and setting up the Open Finance
# Conformance Suite.
setup-dev:
	@go mod download
	@pre-commit install
	@make keys
	@make setup-cs

# Clone and build the Open Finance Conformance Suite.
# Also, generate a configuration file for the suite using files in /keys.
# A configuration file for the conformance suite is also generated based on the files inside /keys.
# Note: The Dockerfile to build the conformance suite jar is missing, then it is
# being added it manually.
setup-cs:
	@if [ ! -d "conformance-suite" ]; then \
	  echo "Cloning open finance conformance suite repository..."; \
	  git clone --branch master --single-branch --depth=1 https://gitlab.com/raidiam-conformance/open-banking/certification.git conformance-suite; \
	fi

	@make build-cs

	@make cs-config

# Runs the main MockBank components.
run:
	@docker-compose --profile main up

# Start MockBank along with the Open Finance Conformance Suite.
run-with-cs:
	@docker-compose --profile main --profile conformance up

# Runs only the MockBank dependencies necessary for local development. With this
# command the MockBank server can run and be debugged in the local host.
run-dev:
	@docker-compose --profile dev up

# Runs the local development environment with both MockBank and the Conformance
# Suite. With this command the MockBank server can run and be debugged in the
# local host with the Conformance Suite.
run-dev-with-cs:
	@docker-compose --profile dev --profile conformance up

# Run the Conformance Suite.
run-cs:
	docker compose --profile conformance up

# Generate certificates, private keys, and JWKS files for both the server and clients.
.PHONY: keys
keys:
	@go run cmd/keymaker/main.go

# Build the MockBank Docker Image.
build-mockbank:
	@docker-compose build mockbank

# Build the Conformance Suite JAR file.
build-cs:
	@docker compose run cs-builder

# Create a Conformance Suite configuration file using the client keys in /keys.
cs-config:
	@jq --arg clientOneCert "$$(<keys/client_one.crt)" \
	   --arg clientOneKey "$$(<keys/client_one.key)" \
	   --arg clientTwoCert "$$(<keys/client_two.crt)" \
	   --arg clientTwoKey "$$(<keys/client_two.key)" \
	   --argjson clientOneJwks "$$(jq . < keys/client_one.jwks)" \
	   --argjson clientTwoJwks "$$(jq . < keys/client_two.jwks)" \
	   '.client.jwks = $$clientOneJwks | \
	    .mtls.cert = $$clientOneCert | \
	    .mtls.key = $$clientOneKey | \
	    .client2.jwks = $$clientTwoJwks | \
	    .mtls2.cert = $$clientTwoCert | \
	    .mtls2.key = $$clientTwoKey' \
		cs_config_base.json > cs_config.json

	@echo "Conformance Suite config successfully written to cs_config.json"
