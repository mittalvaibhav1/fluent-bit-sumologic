NAME = fluent-bit-sumologic
VERSION = latest
MAINFEST = fluent-bit-k8s.yaml

build:
	docker build -t $(NAME):$(VERSION) .

delete:
	kubectl delete -f $(MAINFEST)

load:
	minikube image load --overwrite=true $(NAME):$(VERSION)

apply:
	kubectl apply -f $(MAINFEST)

run: build load apply

drun: build delete load apply
