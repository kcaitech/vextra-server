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
    set VERSION_TAG=latest
)

:: 构建builder镜像：
:: docker build --target builder -t kcserver-builder_image:latest -f ../../%SERVICE_NAME%/Dockerfile ../../

docker build -t %SERVICE_NAME%:%VERSION_TAG% -f ../../%SERVICE_NAME%/Dockerfile ../../
docker tag %SERVICE_NAME%:%VERSION_TAG% registry.protodesign.cn:36000/kcserver/%SERVICE_NAME%:%VERSION_TAG%
docker login registry.protodesign.cn:36000 -u admin -p Kcai1212
docker push registry.protodesign.cn:36000/kcserver/%SERVICE_NAME%:%VERSION_TAG%
