:: 对所有模块执行go mod tidy -v命令

set original_path=%cd%

for %%i in ("%cd%") do set base_dir=%%~ni
if not "%base_dir%"=="kcserver" (
    echo 请在kcserver目录下执行
    cd %original_path%
    exit /b
)
set base_dir=%cd%

call :GoModTidy apigateway
call :GoModTidy authservice
call :GoModTidy documentservice
call :GoModTidy userservice
goto :exit

:exit
cd %original_path%
exit /b

:GoModTidy
    cd %base_dir%\%1
    echo 正在处理：%1
    @echo on
    go mod tidy -v
    @echo off
    exit /b

