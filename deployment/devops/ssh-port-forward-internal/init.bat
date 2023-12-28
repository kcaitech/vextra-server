: 生成密钥并将公钥添加到跳板机上

if not exist "%USERPROFILE%\.ssh\id_rsa.pub" (
    ssh-keygen -t rsa -b 4096
)
scp "%USERPROFILE%\.ssh\id_rsa.pub" root@192.168.0.18:/tmp/id_rsa.pub
scp "%~dp0copy-ssh-key.sh" root@192.168.0.18:/tmp/copy-ssh-key.sh
ssh root@192.168.0.18 "bash /tmp/copy-ssh-key.sh"
