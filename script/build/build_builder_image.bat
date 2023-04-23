:: 构建“构建用镜像：builder_image”
docker build --target builder -t builder_image:latest -f ./apigateway/Dockerfile .
docker build --target builder -t builder_image:latest -f ./authservice/Dockerfile .
docker build --target builder -t builder_image:latest -f ./documentservice/Dockerfile .
docker build --target builder -t builder_image:latest -f ./userservice/Dockerfile .