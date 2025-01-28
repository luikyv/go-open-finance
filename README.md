# go-open-finance
An implementation of the Brazil Open Finance specifications in Go.

## Open API Specs
The following Open Finance Open API Specifications are implemented.

### Phase 2
* [API Consents v3.2.0](https://openbanking-brasil.github.io/openapi/swagger-apis/consents/3.2.0.yml)
* [API Resources v3.0.0](https://openbanking-brasil.github.io/openapi/swagger-apis/resources/3.0.0.yml)
* [API Customers v2.2.0](https://openbanking-brasil.github.io/openapi/swagger-apis/customers/2.2.0.yml)
* [API Accounts v2.4.1](https://openbanking-brasil.github.io/openapi/swagger-apis/accounts/2.4.1.yml)

## Mocked Users
Below is the list of pre-configured users in MockBank. These users are available for testing and interaction within the system.

### Bob
- Username: bob@mail.com
- Password: pass
- CPF: 78628584099

Bob is the main user for MockBank, and most scenarios have been implemented for him.

### Alice
- Username: alice@mail.com
- Password: pass
- CPF: 96362357086

Alice is assigned to a joint account, and her credentials are designed for testing such scenarios.

## Local Setup
To ensure MockBank works correctly in your local environment, you need to update your system's hosts file (usually located at /etc/hosts on Unix-based systems or C:\Windows\System32\drivers\etc\hosts on Windows). This step allows your machine to resolve the required domains for MockBank.
```bash
127.0.0.1 mockbank.local
127.0.0.1 matls-mockbank.local
```

If you're running MockBank directly on your machine instead of in a Docker container, add this entry. It ensures MockBank can resolve the mocked directory served by the NGINX container:
```bash
127.0.0.1 directory
```

If you are developing or modifying this project, start by running `make setup-dev`. For this you will need:
* Docker and Docker Compose installed.
* Go 1.22.x installed and properly configured in the development environment.
* Pre-commit installed for managing Git hooks.
* `jq` installed for JSON processing in the Makefile commands.

Once the setup is complete, you'll be able to use all other make commands.

If you only need to run the project without modifying it, you can use the simpler setup with `make setup`. For this you only need Docker and Docker Compose installed. After this setup, you can start the services using `make run`.

## Dependencies
This project relies significantly on some Go dependencies that streamline development and reduce boilerplate code.

### go-oidc
[go-oidc](https://github.com/luikyv/go-oidc) is a configurable OpenID provider written in Go. It handles OAuth-related functionalities, including authentication, token issuance, and scopes. Familiarity with this library's concepts is important for understanding the project's implementation of these aspects.

## TODOs
* Make mongo db remove expired records.
* Env. Defaults to DEV and log warning?
* Business.
* Add more logs.
* Better way to generate the software statement assertion.
