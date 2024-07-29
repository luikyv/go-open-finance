start:
	@docker-compose up -d
	@cd gopf && go run .

init-keys:
	# Generate the server's key and certificate.
	@openssl genpkey -algorithm RSA -out keys/server_key.pem
	@openssl req -new -key keys/server_key.pem -out keys/signing_req.csr -config keys/template_csr.conf
	@openssl x509 -req -days 365 -in keys/signing_req.csr -signkey keys/server_key.pem -out  keys/server_cert.pem

	# Generate the client CA's key and certificate.
	@openssl genpkey -algorithm RSA -out keys/client_ca_key.pem
	@openssl req -new -key keys/client_ca_key.pem -out keys/signing_req.csr -config keys/template_csr.conf
	@openssl x509 -req -days 365 -in keys/signing_req.csr -signkey keys/client_ca_key.pem -out  keys/client_ca_cert.pem

	# Generate the client one's key and certificate.
	@openssl genpkey -algorithm RSA -out keys/client_one_key.pem
	@openssl req -new -key keys/client_one_key.pem -out keys/signing_req.csr -config keys/template_csr.conf
	@openssl x509 -req -days 365 -in keys/signing_req.csr -signkey keys/client_ca_key.pem -out  keys/client_one_cert.pem

	# Generate the client two's key and certificate.
	@openssl genpkey -algorithm RSA -out keys/client_two_key.pem
	@openssl req -new -key keys/client_two_key.pem -out keys/signing_req.csr -config keys/template_csr.conf
	@openssl x509 -req -days 365 -in keys/signing_req.csr -signkey keys/client_ca_key.pem -out  keys/client_two_cert.pem

	@rm keys/signing_req.csr
