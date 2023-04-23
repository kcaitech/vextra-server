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
call :CreateLog %base_dir%\apigateway
call :CreateLog %base_dir%\authservice
call :CreateLog %base_dir%\documentservice
call :CreateLog %base_dir%\userservice

cd %original_path%
exit /b

:CreateLog
    cd %1
    if not exist log\all.log (
        mkdir log
        type nul > log\all.log
    )
    exit /b
