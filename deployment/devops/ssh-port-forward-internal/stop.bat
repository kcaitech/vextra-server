@echo off
for /F %%i in (ssh_pids.txt) do (
   taskkill /F /PID %%i
)
del ssh_pids.txt
