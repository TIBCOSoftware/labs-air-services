SHELL := /bin/bash
SCRIPTS_PATH      := scripts

.PHONY: build-push-delete-air-flogo-builder
build-push-delete-air-flogo-builder: build-air-flogo-builder push-image delete-local-image

.PHONY: build-air-flogo-builder
build-air-flogo-builder:
	@$(SCRIPTS_PATH)/build_air_flogo_builder.sh ${IMAGE_NAME} ${IMAGE_TAG} ${IMAGE_URL} ${BUILDER_TYPE}

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