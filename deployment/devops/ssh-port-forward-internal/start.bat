:: 设置k8s节点机的ssh端口的转发（通过跳板机）
@echo off
chcp 65001 > nul
setlocal enabledelayedexpansion

call add_ssh_key.bat root@192.168.0.18 22

:: 如果存在ssh_pids文件则执行stop.bat脚本删除之前的ssh进程
:: if exist ssh_pids.txt (
    :: call stop.bat
:: )

:: 定义端口映射规则
:: 节点ssh
set ports[0]=2210:172.16.0.10:22
set ports[1]=2220:172.16.0.20:22
set ports[2]=2221:172.16.0.21:22
set ports[3]=2222:172.16.0.22:22
:: kuboard访问端口
set ports[4]=30001:172.16.0.20:30001
:: apiserver访问端口
set ports[5]=9443:172.16.0.19:9443
:: docker registry
set ports[6]=35000:172.16.0.19:35000
:: docker registry-ui
set ports[7]=35001:172.16.0.19:35001
:: bitnami-mariadb
set ports[8]=30100:172.16.0.19:30100
:: percona-mongodb
set ports[9]=30200:172.16.0.19:30200
:: bitnami-redis
set ports[10]=30300:172.16.0.19:30300
set ports[11]=30500:172.16.0.19:30500
:: documentservice
set ports[12]=32003:172.16.0.19:32003

:: 获取数组长度
set length=0
:loop
if defined ports[%length%] (
    set /a length+=1
    goto loop
)
if %length%==0 (
    echo 未定义端口映射规则
    pause > nul
    exit /b
)

set /a end=length-1
for /L %%i in (0,1,%end%) do (
    set port=!ports[%%i]!
    powershell.exe -NoProfile -Command "$process = Start-Process -FilePath 'ssh' -ArgumentList '-N -L !port! root@192.168.0.18' -NoNewWindow -PassThru; $process.Id | Out-File ssh_pids.txt"
)

echo 脚本已执行...
pause > nul

endlocal
