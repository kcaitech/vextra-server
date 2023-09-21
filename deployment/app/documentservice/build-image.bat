:: 构建镜像并上传到仓库
@echo off
chcp 65001 > nul

set SERVICE_NAME=%1
set VERSION_TAG=%2

if "%SERVICE_NAME%"=="" (
    echo "参数错误：build-image.bat [SERVICE_NAME] [VERSION_TAG]"
    exit /b
)
if "%VERSION_TAG%"=="" (
    echo "参数错误：build-image.bat [SERVICE_NAME] [VERSION_TAG]"
    exit /b
)

docker build -t %SERVICE_NAME%:%VERSION_TAG% -f ../../../%SERVICE_NAME%/Dockerfile ../../../
docker tag %SERVICE_NAME%:%VERSION_TAG% docker-registry.protodesign.cn:35000/%SERVICE_NAME%:%VERSION_TAG%
docker login docker-registry.protodesign.cn:35000 -u kcai -p kcai1212
docker push docker-registry.protodesign.cn:35000/%SERVICE_NAME%:%VERSION_TAG%
