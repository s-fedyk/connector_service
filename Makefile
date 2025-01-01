naive: build
	docker run -p 80:80 pomidoro/connector-service:1
build:
	protoc --go_out=./proto --go-grpc_out=./proto -I. proto/ImageService.proto
	docker build -t pomidoro/connector-service:1 .
	docker push docker.io/pomidoro/connector-service:1
dev: build
	kubectl apply -f k8s/dev/deployment.yaml
	kubectl apply -f k8s/dev/service.yaml
	kubectl apply -f k8s/dev/configmap.yaml
prod: 
	kubectl apply -f k8s/prod/deployment.yaml
	kubectl apply -f k8s/prod/service.yaml
	kubectl apply -f k8s/prod/configmap.yaml
teardown:
	- kubectl delete deployment connector-service-deployment
	- kubectl delete service connector-service
	- kubectl delete configmap service-config
