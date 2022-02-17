SHELL := /bin/bash
SCRIPTS_PATH      := scripts

ifndef IMAGE_NAME
override IMAGE_NAME = labs-air-services
endif
ifndef IMAGE_TAG
override IMAGE_TAG = latest
endif
ifndef ECR_REGISTRY
override ECR_REGISTRY = public.ecr.aws
endif
ifndef ECR_REPO_NAME
override ECR_REPO_NAME = tibcolabs
endif
ifndef IMAGE_URL
override IMAGE_URL = "$(ECR_REGISTRY)/$(ECR_REPO_NAME)"
endif

.PHONY: build-push-delete-air-flogo-builder
build-push-delete-air-flogo-builder: build-air-flogo-builder push-image delete-local-image

.PHONY: build-air-flogo-builder
build-air-flogo-builder:
	@$(SCRIPTS_PATH)/build_air_flogo_builder.sh ${IMAGE_NAME} ${IMAGE_TAG} ${IMAGE_URL} ${BUILDER_TYPE} ${IMAGE_ARCH}

.PHONY: build-push-delete-air-flogo-app
build-push-delete-air-flogo-app: build-air-flogo-app push-image delete-local-image

.PHONY: build-air-flogo-app
build-air-flogo-app:
	@$(SCRIPTS_PATH)/build_air_flogo_app.sh ${IMAGE_NAME} ${IMAGE_TAG} ${IMAGE_URL} ${FLOGO_APP_NAME}

.PHONY: push-image
push-image:
	@$(SCRIPTS_PATH)/push_image.sh ${IMAGE_NAME} ${IMAGE_TAG} ${IMAGE_URL}

.PHONY: delete-local-image
delete-local-image:
	@$(SCRIPTS_PATH)/delete_local_image.sh ${IMAGE_NAME} ${IMAGE_TAG} ${IMAGE_URL}