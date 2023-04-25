:: 首次构建之前执行
@echo off
chcp 65001 > nul

set original_path=%cd%

for %%i in ("%cd%") do set base_dir=%%~ni
if not "%base_dir%"=="kcserver" (
    echo 请在kcserver目录下执行
    cd %original_path%
    exit /b
)
set base_dir=%cd%

:: 创建空日志文件
call :CreateLog apigateway
call :CreateLog authservice
call :CreateLog documentservice
call :CreateLog userservice

:: 创建docker network
docker network create --subnet=172.21.0.0/16 --gateway=172.21.0.1 db_net_1

cd %original_path%
exit /b

:CreateLog
    cd %base_dir%\%1
    if not exist log\all.log (
        mkdir log
        type nul > log\all.log
    )
    exit /b
