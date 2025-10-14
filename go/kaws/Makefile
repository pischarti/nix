# Operator SDK / Kubebuilder compatible Makefile

# Image URL to use all building/pushing image targets
IMG ?= kaws-operator:latest
# AWS region for deployment
AWS_REGION ?= us-east-1

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: fmt vet ## Build kaws binary.
	go build -o kaws main.go

.PHONY: run
run: fmt vet ## Run the operator from your host.
	go run ./main.go operator

.PHONY: docker-build
docker-build: ## Build docker image with the operator.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the operator.
	docker push ${IMG}

##@ Deployment

.PHONY: install
install: ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/crd/eventrecycler.yaml

.PHONY: uninstall
uninstall: ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/crd/eventrecycler.yaml

.PHONY: deploy
deploy: ## Deploy operator to the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/rbac/role.yaml
	kubectl apply -f config/rbac/leader_election_role.yaml
	kubectl apply -f config/manager/deployment.yaml

.PHONY: undeploy
undeploy: ## Undeploy operator from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/manager/deployment.yaml
	kubectl delete -f config/rbac/leader_election_role.yaml --ignore-not-found=true
	kubectl delete -f config/rbac/role.yaml

.PHONY: deploy-sample
deploy-sample: ## Deploy sample EventRecycler CR.
	kubectl apply -f config/samples/eventrecycler_sample.yaml

##@ Complete Setup

.PHONY: setup-operator
setup-operator: docker-build docker-push install deploy deploy-sample ## Complete operator setup (build, push, install, deploy).
	@echo "Operator deployed successfully!"
	@echo "Watch logs with: kubectl logs -n kube-system -l app=kaws-operator -f"

.PHONY: teardown
teardown: undeploy uninstall ## Complete teardown (undeploy and uninstall).
	@echo "Operator removed successfully!"

