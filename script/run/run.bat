:: （重新）启动所有服务
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

if "%1" == "" (
  :: call :StartService docker_compose\mysql
  :: call :StartService docker_compose\minio

  call :RestartService apigateway
  call :RestartService authservice
  call :RestartService documentservice
  call :RestartService userservice

  goto :exit
)

call :RestartService %1
goto :exit

:exit
cd %original_path%
exit /b

:StartService
    cd %base_dir%\%1
    echo 正在处理：%1
    @echo on
    docker-compose up -d
    @echo off
    exit /b

:RestartService
    cd %base_dir%\%1
    echo 正在处理：%1
    @echo on
    docker-compose up -d --build --force-recreate
    @echo off
    exit /b