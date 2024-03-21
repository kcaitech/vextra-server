@echo off
chcp 65001 > nul
setlocal

:: 读取第一第二个参数
set remote_user_host=%1
set remote_port=%2

:: 如果ssh需要密码才能链接则将本地公钥添加到远程主机上
ssh -o BatchMode=yes %remote_user_host% -p %remote_port% "echo" 2>nul
if %ERRORLEVEL% neq 0 (
    if not exist "%USERPROFILE%\.ssh\id_rsa.pub" (
        ssh-keygen -t rsa -b 4096
    )
    type "%USERPROFILE%\.ssh\id_rsa.pub" | ssh %remote_user_host% -p %remote_port% ^
        "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 700 ~/.ssh && chmod 600 ~/.ssh/authorized_keys"
)

endlocal
