.PHONY: all protos secret_key run_client run_game run_agent run_login run_platform

protos:
	cd internal/protocol && make protos

secret_key: secret_key_alg := Ed25519
secret_key:
	echo "Generating $(secret_key_alg) key pair..."
	mkdir -p ./bin/secret_key/$(secret_key_alg)
	openssl genpkey -algorithm $(secret_key_alg) -out ./bin/secret_key/$(secret_key_alg)/priv.pem
	openssl pkey -in ./bin/secret_key/$(secret_key_alg)/priv.pem -pubout -out ./bin/secret_key/$(secret_key_alg)/pub.pem

gen_user_token: issuer := dev
gen_user_token: exp := 24h
gen_user_token:
	go run internal/tools/gen_user_token/main.go \
		-pri-pem ./configs/secret_key/auth_priv.pem \
		-issuer $(issuer) \
		-exp $(exp) \
		-user-info "{ \
			\"uid\": \"yy01\" \
		}"
		
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

run_game: config_path := ./app/game/configs/dev.toml
run_game: server_id := 1
run_game:
	go run github.com/godyy/ggs/app/game \
		-config-path "$(config_path)" \
		-env-server-id "$(server_id)"

run_agent: config_path := ./app/agent/configs/dev.toml
run_agent: server_id := 1
run_agent:
	go run github.com/godyy/ggs/app/agent \
		-config-path "$(config_path)" \
		-env-server-id "$(server_id)"

run_login: config_path := ./app/login/configs/dev.toml
run_login:
	go run github.com/godyy/ggs/app/login \
		-config-path "$(config_path)"

run_platform: config_path := ./app/platform/configs/dev.toml
run_platform:
	go run github.com/godyy/ggs/app/platform \
		-config-path "$(config_path)"
