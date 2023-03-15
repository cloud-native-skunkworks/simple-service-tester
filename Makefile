.PHONY: build install uninstall
IMG="cnskunkworks/simple-service-tester:v1"
RELEASE="simple-service-tester"
build:
	cd src && docker buildx build --push --platform="linux/amd64,linux/arm64" -t ${IMG} .

install:
	kubectl create ns foo || true
	kubectl create ns bar || true
	helm install ${RELEASE} .
uninstall:
	helm uninstall ${RELEASE}