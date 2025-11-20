.PHONY: all protos secret_key

protos:
	cd internal/proto && make protos

secret_key: secret_key_alg := Ed25519
secret_key:
	echo "Generating $(secret_key_alg) key pair..."
	mkdir -p ./bin/secret_key/$(secret_key_alg)
	openssl genpkey -algorithm $(secret_key_alg) -out ./bin/secret_key/$(secret_key_alg)/priv.pem
	openssl pkey -in ./bin/secret_key/$(secret_key_alg)/priv.pem -pubout -out ./bin/secret_key/$(secret_key_alg)/pub.pem

gen_user_token: issuer := dev
gen_user_token: exp := 24h
gen_user_token:
	go run internal/tools/auth/gen_user_token/main.go \
		-pri-pem ./configs/secret_key/auth_priv.pem \
		-issuer $(issuer) \
		-exp $(exp) \
		-user-info "{ \
			\"uid\": \"yy01\" \
		}"
		
create_indexes: mongo_uri:=mongodb://localhost:27017
create_indexes:
	go run internal/tools/db/create_indexes/main.go \
		-mongo-uri "$(mongo_uri)"

run_client: login_url_root := http://localhost:8080/api/v1
run_client: agent_addr := localhost:22001
run_client: uid := yy01
run_client: server_id := 1
run_client:
	go run github.com/godyy/ggs/app/client \
        -login-url-root "$(login_url_root)" \
        -agent-addr "$(agent_addr)" \
        -sign-key-path "./configs/secret_key/auth_priv.pem" \
        -mode client \
        -client-uid "$(uid)" \
        -client-server-id "$(server_id)"
