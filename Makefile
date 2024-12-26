build:
	protoc --go_out=./proto --go-grpc_out=./proto -I. proto/ImageService.proto
