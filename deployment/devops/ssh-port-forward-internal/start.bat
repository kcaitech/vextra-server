:: 设置k8s节点机的ssh端口的转发（通过跳板机）
@echo off
chcp 65001 > nul
setlocal enabledelayedexpansion

:: 如果存在ssh_pids文件则执行stop.bat脚本删除之前的ssh进程
if exist ssh_pids.txt (
    call stop.bat
)

:: 定义端口映射规则
set ports[0]=2210:169.254.231.10:22
set ports[1]=2220:169.254.231.20:22
set ports[2]=2221:169.254.231.21:22
set ports[3]=2222:169.254.231.22:22
:: 获取数组长度
set length=0
:loop
if defined ports[%length%] (
    set /a length+=1
    goto loop
)

set /a end=length-1
for /L %%i in (0,1,%end%) do (
    set port=!ports[%%i]!
    powershell.exe -NoProfile -Command "$process = Start-Process -FilePath 'ssh' -ArgumentList '-N -L !port! root@192.168.0.18' -NoNewWindow -PassThru; $process.Id | Out-File -Append ssh_pids.txt"
)

echo 脚本已执行...
pause > nul

endlocal
