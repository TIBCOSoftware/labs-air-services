#!/bin/bash

. scripts/tools.sh

image_name=$1
image_tag=$2
image_url=$3
app_name=$4


local_image_name="flogo/${app_name}"
local_image_tag="local_image_tag"


pushd ./flogo > /dev/null


docker build --no-cache --build-arg APP_NAME=${app_name}.json -t ${local_image_name}:${local_image_tag} -f "Dockerfile_flogo_app_builder" .


# Tag image
tag_image $local_image_name $local_image_tag $image_name $image_tag $image_url
