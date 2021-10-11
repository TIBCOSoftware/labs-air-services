#!/bin/bash

. scripts/tools.sh

image_name=$1
image_tag=$2
image_url=$3
flogo_build_type=$4

pushd ./flogo-oss > /dev/null

build_image $image_name "local_image_tag" $image_url "Dockerfile_flogo_builder_${flogo_build_type}"

# Tag image
tag_image $image_name "local_image_tag" $image_name $image_tag $image_url
