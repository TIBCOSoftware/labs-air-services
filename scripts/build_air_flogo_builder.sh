#!/bin/bash

. scripts/tools.sh

image_name=$1
image_tag=$2
image_url=$3
flogo_build_type=$4
image_arch=$5

pushd ./flogo > /dev/null

if [ "$image_arch" -= "amd64" ]; then
    build_image $image_name "local_image_tag" $image_url "Dockerfile_flogo_builder_${flogo_build_type}"
elif
    build_image_platform $image_name "local_image_tag" $image_url "Dockerfile_flogo_builder_${flogo_build_type}" $image_arch
fi

# Tag image
tag_image $image_name "local_image_tag" $image_name $image_tag $image_url
