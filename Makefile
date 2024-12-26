run: build
	docker run -p 80:80 pomidoro/connector-service:1
build:
	protoc --go_out=./proto --go-grpc_out=./proto -I. proto/ImageService.proto
	docker build -t pomidoro/connector-service:1 .
	docker push docker.io/pomidoro/connector-service:1
kube: build
	kubectl apply -f deployment.yaml
	kubectl apply -f service.yaml
teardown:
	- kubectl delete deployment connector-service-deployment
	- kubectl delete service connector-service

