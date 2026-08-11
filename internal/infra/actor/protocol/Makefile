.PHONY: all protos

protos:
	@echo "generate protos"
	find ./protos -name "*.proto" -exec clang-format -i {} \;
	rm -rf ./pb && mkdir ./pb
	protoc -I=./protos --go_out=./pb --go_opt=paths=source_relative ./protos/**/*.proto
	go run tools/gen_register/main.go -proto-path=. && goimports -w registry/.
