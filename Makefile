# Copyright AppsCode Inc. and Contributors
#
# Licensed under the AppsCode Free Trial License 1.0.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Free-Trial-1.0.0.md
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Copyright 2019 AppsCode Inc.
# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

SHELL=/bin/bash -o pipefail

PRODUCT_OWNER_NAME := appscode
PRODUCT_NAME       := kubestash
ENFORCE_LICENSE    ?=

GO_PKG   := stash.appscode.dev
REPO     := $(notdir $(shell pwd))
BIN      := kubestash
COMPRESS ?= no

# Where to push the docker image.
REGISTRY ?= stashed


# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS          ?= "crd:generateEmbeddedObjectMeta=true"
CODE_GENERATOR_IMAGE ?= appscode/gengo:release-1.25
API_GROUPS           ?= addons:v1alpha1 core:v1alpha1 storage:v1alpha1 config:v1alpha1

# This version-strategy uses git tags to set the version string
git_branch       := $(shell git rev-parse --abbrev-ref HEAD)
git_tag          := $(shell git describe --exact-match --abbrev=0 2>/dev/null || echo "")
commit_hash      := $(shell git rev-parse --verify HEAD)
commit_timestamp := $(shell date --date="@$$(git show -s --format=%ct)" --utc +%FT%T)

VERSION          := $(shell git describe --tags --always --dirty)
version_strategy := commit_hash
ifdef git_tag
	VERSION := $(git_tag)
	version_strategy := tag
else
	ifeq (,$(findstring $(git_branch),master HEAD))
		ifneq (,$(patsubst release-%,,$(git_branch)))
			VERSION := $(git_branch)
			version_strategy := branch
		endif
	endif
endif

RESTIC_VER       := 0.13.1

###
### These variables should not need tweaking.
###

SRC_PKGS := apis controllers crds
SRC_DIRS := $(SRC_PKGS) hack/gencrd hack/kubestash-crd-installer# directories which hold app source (not vendored)

DOCKER_PLATFORMS := linux/amd64 linux/arm linux/arm64
BIN_PLATFORMS    := $(DOCKER_PLATFORMS) windows/amd64 darwin/amd64

# Used internally.  Users should pass GOOS and/or GOARCH.
OS   := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))

BASEIMAGE_PROD   ?= gcr.io/distroless/static-debian11
BASEIMAGE_DBG    ?= debian:bullseye

IMAGE            := $(REGISTRY)/$(BIN)
VERSION_PROD     := $(VERSION)
VERSION_DBG      := $(VERSION)-dbg
TAG              := $(VERSION)_$(OS)_$(ARCH)
TAG_PROD         := $(TAG)
TAG_DBG          := $(VERSION)-dbg_$(OS)_$(ARCH)

GO_VERSION       ?= 1.19
BUILD_IMAGE      ?= appscode/golang-dev:$(GO_VERSION)
TEST_IMAGE       ?= appscode/golang-dev:$(GO_VERSION)-stash

OUTBIN = bin/$(OS)_$(ARCH)/$(BIN)
ifeq ($(OS),windows)
  OUTBIN = bin/$(OS)_$(ARCH)/$(BIN).exe
endif

# Directories that we need created to build/tests.
BUILD_DIRS  := bin/$(OS)_$(ARCH)     \
               .go/bin/$(OS)_$(ARCH) \
               .go/cache             \
               hack/config           \
               $(HOME)/.credentials  \
               $(HOME)/.kube         \
               $(HOME)/.minikube

DOCKERFILE_PROD  = Dockerfile.in
DOCKERFILE_DBG   = Dockerfile.dbg
DOCKERFILE_TEST  = Dockerfile.test

DOCKER_REPO_ROOT := /go/src/$(GO_PKG)/$(REPO)

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

version: ## Print the version information resolved by this Makefile.
	@echo ::set-output name=version::$(VERSION)
	@echo ::set-output name=version_strategy::$(version_strategy)
	@echo ::set-output name=git_tag::$(git_tag)
	@echo ::set-output name=git_branch::$(git_branch)
	@echo ::set-output name=commit_hash::$(commit_hash)
	@echo ::set-output name=commit_timestamp::$(commit_timestamp)

.PHONY: clean
clean: ## Cleanup build cache and other temporary files.
	rm -rf .go bin

##@ Generators

.PHONY: gen
gen: manifests openapi ## Runs all the generators. Run this command after modifying the type files or any "+kubebuilder" marker.

.PHONY: manifests
manifests: gen-deepcopy gen-crds label-crds gen-rbac gen-webhook ## Runs all the generator except "openapi".

.PHONY: generate
generate: manifests

.PHONY: gen-deepcopy
gen-deepcopy: ## Generate DeepCopy functions for the APIs. Run this command after modifying the API types
	@echo "Generating DeepCopy........."
	@docker run --rm								\
		-u $$(id -u):$$(id -g)						\
		-v /tmp:/.cache								\
		-v $$(pwd):$(DOCKER_REPO_ROOT)				\
		-w $(DOCKER_REPO_ROOT)						\
	    --env HTTP_PROXY=$(HTTP_PROXY)				\
	    --env HTTPS_PROXY=$(HTTPS_PROXY)			\
		$(CODE_GENERATOR_IMAGE)						\
		controller-gen								\
			object:headerFile="hack/license/go.txt"	\
			paths="./..."

.PHONY: gen-rbac
gen-rbac: ## Generate RBAC resources from the "+kubebuilder:rbac" markers. Run this command after adding/removing/updating any "+kubebuilder:rbac" marker in the controller code.
	@echo "Generating RBAC resources........"
	@docker run --rm 	                    \
		-u $$(id -u):$$(id -g)              \
		-v /tmp:/.cache                     \
		-v $$(pwd):$(DOCKER_REPO_ROOT)      \
		-w $(DOCKER_REPO_ROOT)              \
	    --env HTTP_PROXY=$(HTTP_PROXY)      \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)    \
		$(CODE_GENERATOR_IMAGE)             \
		controller-gen                      \
			rbac:roleName=kubestash 		\
			paths="./..."             	    \
			output:rbac:artifacts:config=config/rbac

.PHONY: gen-webhook
gen-webhook: ## Generate the WebhookConfiguration files. Run this command after adding/updating any webhook.
	@echo "Generating WebhookConfigurations........."
	@docker run --rm 	                    \
		-u $$(id -u):$$(id -g)              \
		-v /tmp:/.cache                     \
		-v $$(pwd):$(DOCKER_REPO_ROOT)      \
		-w $(DOCKER_REPO_ROOT)              \
	    --env HTTP_PROXY=$(HTTP_PROXY)      \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)    \
		$(CODE_GENERATOR_IMAGE)             \
		controller-gen                      \
			webhook							\
			paths="./..." 					\
			output:webhook:artifacts:config=config/webhook

# Generate CRD manifests
.PHONY: gen-crds
gen-crds: ## Generate CRDs YAMLs from the API types. Run this command after you modify the API types file.
	@echo "Generating CRD manifests........."
	@docker run --rm 	                    \
		-u $$(id -u):$$(id -g)              \
		-v /tmp:/.cache                     \
		-v $$(pwd):$(DOCKER_REPO_ROOT)      \
		-w $(DOCKER_REPO_ROOT)              \
	    --env HTTP_PROXY=$(HTTP_PROXY)      \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)    \
		$(CODE_GENERATOR_IMAGE)             \
		controller-gen                      \
			$(CRD_OPTIONS)                  \
			paths="./..."					\
			output:crd:artifacts:config=crds

.PHONY: label-crds
label-crds: $(BUILD_DIRS)
	@for f in crds/*.yaml; do \
		echo "applying app.kubernetes.io/name=kubestash label to $$f"; \
		kubectl label --overwrite -f $$f --local=true -o yaml app.kubernetes.io/name=kubestash > bin/crd.yaml; \
		mv bin/crd.yaml $$f; \
	done
	@echo ""

# Generate openapi schema
OPENAPI_VERSION	:= v0.1.0
latest_tag := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "")
ifdef latest_tag
	OPENAPI_VERSION = $(latest_tag)
endif
.PHONY: openapi
openapi: $(addprefix openapi-, $(subst :,_, $(API_GROUPS))) ## Generate OpenAPI swagger.json file for the APIs. Run this command after you modify the API types file.
	@echo "Generating openapi/swagger.json........"
	@docker run --rm 	                                 \
		-u $$(id -u):$$(id -g)                           \
		-v /tmp:/.cache                                  \
		-v $$(pwd):$(DOCKER_REPO_ROOT)                   \
		-w $(DOCKER_REPO_ROOT)                           \
		--env HTTP_PROXY=$(HTTP_PROXY)                   \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                 \
		--env                             \
		--env GOFLAGS="-mod=vendor"                      \
		$(BUILD_IMAGE)                                   \
		go run hack/gencrd/main.go --version=$(OPENAPI_VERSION)

openapi-%:
	@echo "Generating openapi schema for $(subst _,/,$*)"
	@mkdir -p .config/api-rules
	@docker run --rm 	                                 \
		-u $$(id -u):$$(id -g)                           \
		-v /tmp:/.cache                                  \
		-v $$(pwd):$(DOCKER_REPO_ROOT)                   \
		-w $(DOCKER_REPO_ROOT)                           \
		--env HTTP_PROXY=$(HTTP_PROXY)                   \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                 \
		$(CODE_GENERATOR_IMAGE)                          \
		openapi-gen                                      \
			--v 1 --logtostderr                          \
			--go-header-file "./hack/license/go.txt" \
			--input-dirs "$(GO_PKG)/$(REPO)/apis/$(subst _,/,$*),k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/runtime,k8s.io/apimachinery/pkg/util/intstr,k8s.io/apimachinery/pkg/version,k8s.io/api/core/v1,k8s.io/api/apps/v1,kmodules.xyz/offshoot-api/api/v1,kmodules.xyz/custom-resources/apis/appcatalog/v1alpha1,k8s.io/api/rbac/v1,kmodules.xyz/objectstore-api/api/v1,kmodules.xyz/prober/api/v1,kmodules.xyz/client-go/api/v1,stash.appscode.dev/kubestash/apis" \
			--output-package "$(GO_PKG)/$(REPO)/apis/$(subst _,/,$*)" \
			--report-filename .config/api-rules/violation_exceptions.list


##@ Build

# If you want to build all binaries, see the 'all-build' rule.
# If you want to build all containers, see the 'all-container' rule.
# If you want to build AND push all containers, see the 'all-push' rule.
all: fmt build ## Format and compile code.

# For the following OS/ARCH expansions, we transform OS/ARCH into OS_ARCH
# because make pattern rules don't match with embedded '/' characters.

build-%:
	@$(MAKE) build                        \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

all-build: $(addprefix build-, $(subst /,_, $(BIN_PLATFORMS)))


fmt: $(BUILD_DIRS) ## Format source code.
	@docker run                                                 \
	    -i                                                      \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    $(BUILD_IMAGE)                                          \
	    /bin/bash -c "                                          \
	        REPO_PKG=$(GO_PKG)                                  \
	        ./hack/fmt.sh $(SRC_DIRS)                           \
	    "

build: $(OUTBIN) ## Compile source code and build binary.

# The following structure defeats Go's (intentional) behavior to always touch
# result files, even if they have not changed.  This will still run `go` but
# will not trigger further work if nothing has actually changed.

$(OUTBIN): .go/$(OUTBIN).stamp
	@true

# This will build the binary under ./.go and update the real binary iff needed.
.PHONY: .go/$(OUTBIN).stamp
.go/$(OUTBIN).stamp: $(BUILD_DIRS)
	@echo "making $(OUTBIN)"
	@docker run                                                 \
	    -i                                                      \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    $(BUILD_IMAGE)                                          \
	    /bin/bash -c "                                          \
	        PRODUCT_OWNER_NAME=$(PRODUCT_OWNER_NAME)            \
	        PRODUCT_NAME=$(PRODUCT_NAME)                        \
	        ENFORCE_LICENSE=$(ENFORCE_LICENSE)                  \
	        ARCH=$(ARCH)                                        \
	        OS=$(OS)                                            \
	        VERSION=$(VERSION)                                  \
	        version_strategy=$(version_strategy)                \
	        git_branch=$(git_branch)                            \
	        git_tag=$(git_tag)                                  \
	        commit_hash=$(commit_hash)                          \
	        commit_timestamp=$(commit_timestamp)                \
	        ./hack/build.sh                                     \
	    "
	@if [ $(COMPRESS) = yes ] && [ $(OS) != darwin ]; then          \
		echo "compressing $(OUTBIN)";                               \
		@docker run                                                 \
		    -i                                                      \
		    --rm                                                    \
		    -u $$(id -u):$$(id -g)                                  \
		    -v $$(pwd):/src                                         \
		    -w /src                                                 \
		    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
		    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
		    -v $$(pwd)/.go/cache:/.cache                            \
		    --env HTTP_PROXY=$(HTTP_PROXY)                          \
		    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
		    $(BUILD_IMAGE)                                          \
		    upx --brute /go/$(OUTBIN);                              \
	fi
	@if ! cmp -s .go/$(OUTBIN) $(OUTBIN); then \
	    mv .go/$(OUTBIN) $(OUTBIN);            \
	    date >$@;                              \
	fi
	@echo

ADDTL_LINTERS   := goconst,gofmt,goimports,unparam

.PHONY: lint
lint: $(BUILD_DIRS) ## Run linter.
	@echo "running linter"
	@docker run                                                 \
	    -i                                                      \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    --env                                    \
	    --env GOFLAGS="-mod=vendor"                             \
	    $(BUILD_IMAGE)                                          \
        golangci-lint run --enable $(ADDTL_LINTERS) --timeout=10m --skip-files="generated.*\.go$\" --skip-dirs-use-default --skip-dirs=client,vendor

$(BUILD_DIRS):
	@mkdir -p $@

.PHONY: add-license
add-license: ## Add license header to the source files.
	@echo "Adding license header"
	@docker run --rm 	                                 \
		-u $$(id -u):$$(id -g)                           \
		-v /tmp:/.cache                                  \
		-v $$(pwd):$(DOCKER_REPO_ROOT)                   \
		-w $(DOCKER_REPO_ROOT)                           \
		--env HTTP_PROXY=$(HTTP_PROXY)                   \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                 \
		$(BUILD_IMAGE)                                   \
		ltag -t "./hack/license" --excludes "vendor contrib libbuild" -v

.PHONY: check-license
check-license: ## Check if license header has been added to all the source files.
	@echo "Checking files for license header"
	@docker run --rm 	                                 \
		-u $$(id -u):$$(id -g)                           \
		-v /tmp:/.cache                                  \
		-v $$(pwd):$(DOCKER_REPO_ROOT)                   \
		-w $(DOCKER_REPO_ROOT)                           \
		--env HTTP_PROXY=$(HTTP_PROXY)                   \
		--env HTTPS_PROXY=$(HTTPS_PROXY)                 \
		$(BUILD_IMAGE)                                   \
		ltag -t "./hack/license" --excludes "vendor contrib libbuild" --check -v

##@ Docker

all-container: $(addprefix container-, $(subst /,_, $(DOCKER_PLATFORMS)))

all-push: $(addprefix push-, $(subst /,_, $(DOCKER_PLATFORMS)))

container-%:
	@$(MAKE) container                    \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

push-%:
	@$(MAKE) push                         \
	    --no-print-directory              \
	    GOOS=$(firstword $(subst _, ,$*)) \
	    GOARCH=$(lastword $(subst _, ,$*))

# Used to track state in hidden files.
DOTFILE_IMAGE    = $(subst /,_,$(IMAGE))-$(TAG)

container: bin/.container-$(DOTFILE_IMAGE)-PROD bin/.container-$(DOTFILE_IMAGE)-DBG ## Make docker image from the binary.
bin/.container-$(DOTFILE_IMAGE)-%: bin/$(OS)_$(ARCH)/$(BIN) $(DOCKERFILE_%)
	@echo "container: $(IMAGE):$(TAG_$*)"
	@sed                                            \
	    -e 's|{ARG_BIN}|$(BIN)|g'                   \
	    -e 's|{ARG_ARCH}|$(ARCH)|g'                 \
	    -e 's|{ARG_OS}|$(OS)|g'                     \
	    -e 's|{ARG_FROM}|$(BASEIMAGE_$*)|g'         \
	    -e 's|{RESTIC_VER}|$(RESTIC_VER)|g'         \
	    $(DOCKERFILE_$*) > bin/.dockerfile-$*-$(OS)_$(ARCH)
	@DOCKER_CLI_EXPERIMENTAL=enabled docker buildx build --platform $(OS)/$(ARCH) --load --pull -t $(IMAGE):$(TAG_$*) -f bin/.dockerfile-$*-$(OS)_$(ARCH) .
	@docker images -q $(IMAGE):$(TAG_$*) > $@
	@echo

push: bin/.push-$(DOTFILE_IMAGE)-PROD bin/.push-$(DOTFILE_IMAGE)-DBG ## Make docker image and push into DockerHub.
bin/.push-$(DOTFILE_IMAGE)-%: bin/.container-$(DOTFILE_IMAGE)-%
	@docker push $(IMAGE):$(TAG_$*)
	@echo "pushed: $(IMAGE):$(TAG_$*)"
	@echo

.PHONY: push-to-kind
push-to-kind: container ## Build docker image and push into local Kind cluster.
	@echo "Loading docker image into kind cluster...."
	@kind load docker-image $(IMAGE):$(TAG)
	@echo "Image has been pushed successfully into kind cluster."

CRD_INSTALLER_TAG ?=$(TAG)
KO := $(shell go env GOPATH)/bin/ko

.PHONY: push-crd-installer
push-crd-installer: $(BUILD_DIRS) install-ko ## Build and push CRD installer image
	@echo "Pushing CRD installer image....."
	DOCKER_CLI_EXPERIMENTAL=enabled KO_DOCKER_REPO=$(REGISTRY) $(KO) publish ./hack/kubestash-crd-installer --tags $(CRD_INSTALLER_TAG)  --base-import-paths  --platform=all

.PHONY: docker-manifest
docker-manifest: docker-manifest-PROD docker-manifest-DBG ## Make docker manifest for multi-arch docker images.
docker-manifest-%:
	DOCKER_CLI_EXPERIMENTAL=enabled docker manifest create -a $(IMAGE):$(VERSION_$*) $(foreach PLATFORM,$(DOCKER_PLATFORMS),$(IMAGE):$(VERSION_$*)_$(subst /,_,$(PLATFORM)))
	DOCKER_CLI_EXPERIMENTAL=enabled docker manifest push $(IMAGE):$(VERSION_$*)

##@ Deploy
REGISTRY_SECRET 		?=
OPERATOR_NAMESPACE		?= kubestash
LICENSE_FILE    		?=
IMAGE_PULL_POLICY 		?=IfNotPresent

ifeq ($(strip $(REGISTRY_SECRET)),)
	IMAGE_PULL_SECRETS =
else
	IMAGE_PULL_SECRETS = --set imagePullSecrets[0].name=$(REGISTRY_SECRET)
endif

.PHONY: run
run: ## Run the operator locally.
	go run -mod=vendor *.go run \
		--v=3 \
		--secure-port=8443 \
		--kubeconfig=$(KUBECONFIG) \
		--authorization-kubeconfig=$(KUBECONFIG) \
		--authentication-kubeconfig=$(KUBECONFIG) \
		--authentication-skip-lookup \
		--docker-registry=$(REGISTRY) \
		--image-tag=$(TAG)

.PHONY: install
install: ## Install KubeStash in the current cluster.
	@cd ../installer; \
	helm dependency update charts/kubestash ;                   		\
	helm upgrade -i kubestash charts/kubestash --wait --create-namespace	\
		--namespace=$(OPERATOR_NAMESPACE)								\
		--set-file global.license=$(LICENSE_FILE)						\
		--set kubestash-operator.operator.registry=$(REGISTRY)			\
		--set kubestash-operator.operator.tag=$(TAG)	   				\
		--set kubestash-operator.imagePullPolicy=$(IMAGE_PULL_POLICY)	\
		--set kubestash-operator.crdInstaller.tag=$(CRD_INSTALLER_TAG) 	\
		$(IMAGE_PULL_SECRETS);				                			\
	kubectl wait --for=condition=Ready pods -l 'app.kubernetes.io/name=kubestash-operator,app.kubernetes.io/instance=kubestash' --timeout=5m -n $(OPERATOR_NAMESPACE)

.PHONY: uninstall
uninstall: ## Uninstall KubeSash from the current cluster. This will not remove the registered CRDs.
	@cd ../installer; \
	helm uninstall kubestash --namespace=$(OPERATOR_NAMESPACE) || true

.PHONY: deploy-to-kind
deploy-to-kind: uninstall push-to-kind install ## Build and deploy the operator in the local Kind cluster.

.PHONY: purge
purge: uninstall ## Uninstall KubeStash and remove the registered CRDs from the current cluster.
	kubectl delete crds -l app.kubernetes.io/name=kubestash

##@ Testing

.PHONY: test
test: unit-tests e2e-tests ## Run both unit tests and E2E tests.

bin/.container-$(DOTFILE_IMAGE)-TEST:
	@echo "container: $(TEST_IMAGE)"
	@sed                                            \
	    -e 's|{ARG_BIN}|$(BIN)|g'                   \
	    -e 's|{ARG_ARCH}|$(ARCH)|g'                 \
	    -e 's|{ARG_OS}|$(OS)|g'                     \
	    -e 's|{ARG_FROM}|$(BUILD_IMAGE)|g'          \
	    -e 's|{RESTIC_VER}|$(RESTIC_VER)|g'         \
	    $(DOCKERFILE_TEST) > bin/.dockerfile-TEST-$(OS)_$(ARCH)
	@DOCKER_CLI_EXPERIMENTAL=enabled docker buildx build --platform $(OS)/$(ARCH) --load --pull -t $(TEST_IMAGE) -f bin/.dockerfile-TEST-$(OS)_$(ARCH) .
	@docker images -q $(TEST_IMAGE) > $@
	@echo

unit-tests: $(BUILD_DIRS) bin/.container-$(DOTFILE_IMAGE)-TEST ## Run unit tests only.
	@docker run                                                 \
	    -i                                                      \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    $(TEST_IMAGE)                                           \
	    /bin/bash -c "                                          \
	        ARCH=$(ARCH)                                        \
	        OS=$(OS)                                            \
	        VERSION=$(VERSION)                                  \
	        ./hack/test.sh $(SRC_PKGS)                          \
	    "

# - e2e-tests can hold both ginkgo args (as GINKGO_ARGS) and program/tests args (as TEST_ARGS).
#       make e2e-tests TEST_ARGS="--selfhosted-operator=false --storageclass=standard" GINKGO_ARGS="--flakeAttempts=2"
#
# - Minimalist:
#       make e2e-tests
#
# NB: -t is used to catch ctrl-c interrupt from keyboard and -t will be problematic for CI.

GINKGO_ARGS ?= "--flakeAttempts=2"
TEST_ARGS   ?=

.PHONY: e2e-tests
e2e-tests: $(BUILD_DIRS) ## Run E2E test in sequential order (only one test will run at a time). You can pass additional argument to Ginkgo using "GINKGO_ARGS" variable. For example: make e2e-tests GINKGO_ARGS="--flakeAttempts=1".
	@docker run                                                 \
	    -i                                                      \
	    --rm                                                    \
	    -u $$(id -u):$$(id -g)                                  \
	    -v $$(pwd):/src                                         \
	    -w /src                                                 \
	    --net=host                                              \
	    -v $(HOME)/.kube:/.kube                                 \
	    -v $(HOME)/.minikube:$(HOME)/.minikube                  \
	    -v $(HOME)/.credentials:$(HOME)/.credentials            \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin                \
	    -v $$(pwd)/.go/bin/$(OS)_$(ARCH):/go/bin/$(OS)_$(ARCH)  \
	    -v $$(pwd)/.go/cache:/.cache                            \
	    --env HTTP_PROXY=$(HTTP_PROXY)                          \
	    --env HTTPS_PROXY=$(HTTPS_PROXY)                        \
	    --env KUBECONFIG=$(KUBECONFIG)                          \
	    --env-file=$$(pwd)/hack/config/.env                     \
	    $(BUILD_IMAGE)                                          \
	    /bin/bash -c "                                          \
	        ARCH=$(ARCH)                                        \
	        OS=$(OS)                                            \
	        VERSION=$(VERSION)                                  \
	        DOCKER_REGISTRY=$(REGISTRY)                         \
	        TAG=$(TAG)                                          \
	        KUBECONFIG=$${KUBECONFIG#$(HOME)}                   \
	        GINKGO_ARGS='$(GINKGO_ARGS)'                        \
	        TEST_ARGS='$(TEST_ARGS) --image-tag=$(TAG)'         \
	        ./hack/e2e.sh                                       \
	    "

.PHONY: e2e-parallel
e2e-parallel: ## Run E2E tests in parallel. Control number of concurrent tests as follows: make e2e-parallel GINKGO_ARGS="--nodes=4".
	@$(MAKE) e2e-tests GINKGO_ARGS="$(GINKGO_ARGS) -p -stream" --no-print-directory

##@ Dev and CI

.PHONY: dev
dev: gen fmt push ## Run generator, format code and build and push docker image into DockerHub.

.PHONY: ci
ci: verify check-license lint build unit-tests ## Run CI checks.

.PHONY: qa
qa: ## Build QA image and push into DockerHub.
	@if [ "$$APPSCODE_ENV" = "prod" ]; then                                              \
		echo "Nothing to do in prod env. Are you trying to 'release' binaries to prod?"; \
		exit 1;                                                                          \
	fi
	@if [ "$(version_strategy)" = "tag" ]; then               \
		echo "Are you trying to 'release' binaries to prod?"; \
		exit 1;                                               \
	fi
	@$(MAKE) clean all-push docker-manifest --no-print-directory

.PHONY: release
release: ## Release final production docker image and push into the DockerHub.
	@if [ "$$APPSCODE_ENV" != "prod" ]; then      \
		echo "'release' only works in PROD env."; \
		exit 1;                                   \
	fi
	@if [ "$(version_strategy)" != "tag" ]; then                    \
		echo "apply tag to release binaries and/or docker images."; \
		exit 1;                                                     \
	fi
	@$(MAKE) clean all-push docker-manifest push-crd-installer --no-print-directory

.PHONY: verify
verify: verify-gen verify-modules ## Verify that the generated codes and modules are up-to-date.

.PHONY: verify-gen
verify-gen: gen fmt ## Verify that the generated files are up-to-date.
	@if !(git diff --exit-code HEAD); then \
		echo "generated files are out of date, run make gen"; exit 1; \
	fi

.PHONY: verify-modules
verify-modules: ## Verify that module files are up-to-date.
	go mod tidy
	go mod vendor
	@if !(git diff --exit-code HEAD); then \
		echo "go module files are out of date"; exit 1; \
	fi

.PHONY: install-ko
install-ko:
	@echo "Installing: github.com/google/ko"
	go install github.com/google/ko@latest
