:: 构建镜像并上传到仓库
@echo off
chcp 65001 > nul

set VERSION_TAG=%1
if "%VERSION_TAG%"=="" (
    set VERSION_TAG=latest
)

call .\build-image.bat apigateway %VERSION_TAG%
call .\build-image.bat authservice %VERSION_TAG%
call .\build-image.bat documentservice %VERSION_TAG%
call .\build-image.bat userservice %VERSION_TAG%
