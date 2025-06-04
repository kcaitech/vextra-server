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

1. 在运行迁移之前，请确保：
   - 源数据库和目标数据库都已创建
   - 目标数据库中的表结构已经存在
   - 有足够的权限访问两个数据库

2. 需要将旧的版本更新服务跑起来
   - branch_migrate

