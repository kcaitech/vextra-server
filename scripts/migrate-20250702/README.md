## 运行迁移

```bash
go run scripts/migrate_db.go
```

## 数据库配置

迁移脚本需要以下数据库配置：

1. 源数据库 (Source)
   - MySQL
     - 用户: root
     - 密码: kKEIjksvnOOIjdZ6rtzE
     - 主机: localhost
     - 端口: 3806
     - 数据库: kcserver
   - MongoDB
     - URL: mongodb://root:jKKsinkjilKKLSW@localhost:28017
     - 数据库: kcserver

2. 目标数据库 (Target)
   - MySQL
     - 用户: root
     - 密码: kKEIjksvnOOIjdZ6rtzE
     - 主机: localhost
     - 端口: 3306
     - 数据库: kcserver
   - MongoDB
     - URL: mongodb://root:jKKsinkjilKKLSW@localhost:27017
     - 数据库: kcserver

3. 用户数据库 (Auth)
   - MySQL
     - 用户: root
     - 密码: password
     - 主机: localhost
     - 端口: 3306
     - 数据库: kcauth

## 注意事项

1. 需要将旧的版本更新服务跑起来
   - branch_migrate
   - package.json @kcdesign/data：1.1.18

2. 旧数据库启动测试：
   - kcdeploy 的 branch_migrate分支
   - 配置文件中version_server的url为 http://192.168.0.131:8088/generate
   - 192.168.0.131 为本机ip地址
   - 8088 旧版版本服务端口
