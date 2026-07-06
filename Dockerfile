FROM golang:1.25-alpine AS builder

WORKDIR /src

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG APP_ENV=dev

ENV CGO_ENABLED=0
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}

COPY . .

# 使用 vendor 目录进行离线构建，只保留编译缓存挂载以加速重复构建。
RUN --mount=type=cache,target=/root/.cache/go-build \
	set -eux; \
	mkdir -p /out; \
	for app in agent game login platform; do \
		test -f "/src/app/${app}/configs/${APP_ENV}.toml"; \
		mkdir -p "/out/configs/${app}"; \
		go build \
			-mod=vendor \
			-trimpath \
			-buildvcs=false \
			-tags "netgo,osusergo" \
			-ldflags="-s -w -buildid=" \
			-o "/out/${app}" \
			"./app/${app}"; \
		cp "/src/app/${app}/configs/${APP_ENV}.toml" "/out/configs/${app}/config.toml"; \
	done

FROM alpine:3.21 AS runtime-base

RUN apk add --no-cache ca-certificates \
	&& addgroup -S -g 65532 app \
	&& adduser -S -D -H -u 65532 -G app app

WORKDIR /app

FROM runtime-base AS agent

COPY --from=builder /out/agent /app/agent
COPY --from=builder /out/configs/agent/config.toml /app/configs/config.toml
COPY configs/secret_key/auth_pub.pem /app/configs/secret_key/auth_pub.pem
WORKDIR /app

RUN chown -R app:app /app

USER app:app

ENTRYPOINT ["./agent", "-config-path", "./configs/config.toml"]

FROM runtime-base AS game

COPY --from=builder /out/game /app/game
COPY --from=builder /out/configs/game/config.toml /app/configs/config.toml
WORKDIR /app

RUN chown -R app:app /app

USER app:app

ENTRYPOINT ["./game", "-config-path", "./configs/config.toml"]

FROM runtime-base AS login

COPY --from=builder /out/login /app/login
COPY --from=builder /out/configs/login/config.toml /app/configs/config.toml
COPY configs/secret_key/auth_pub.pem /app/configs/secret_key/auth_pub.pem
COPY configs/secret_key/auth_priv.pem /app/configs/secret_key/auth_priv.pem
WORKDIR /app

RUN chown -R app:app /app

USER app:app

ENTRYPOINT ["./login", "-config-path", "./configs/config.toml"]

FROM runtime-base AS platform

COPY --from=builder /out/platform /app/platform
COPY --from=builder /out/configs/platform/config.toml /app/configs/config.toml
WORKDIR /app

RUN chown -R app:app /app

USER app:app

ENTRYPOINT ["./platform", "-config-path", "./configs/config.toml"]
