:: 构建“构建用镜像：kcserver-builder_image”
for %%i in ("%cd%") do set base_dir=%%~ni
if not "%base_dir%"=="kcserver" (
    echo "请在kcserver目录下执行"
    cd %original_path%
    exit /b
)

docker build --target builder -t kcserver-builder_image:latest -f ./apigateway/Dockerfile .
docker build --target builder -t kcserver-builder_image:latest -f ./documentservice/Dockerfile .
docker build --target builder -t kcserver-builder_image:latest -f ./userservice/Dockerfile .
docker build --target builder -t kcserver-builder_image:latest -f ./authservice/Dockerfile .