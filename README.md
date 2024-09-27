# kcserver

## 模块介绍
middlewares，鉴权中间件等
api，对外服务api
config，服务相关配置
controllers，业务代码

### config
- `server.port`：服务监听的端口
- `db.dsn`：数据库连接串


# TODO
1. 前端先不使用communication上传文档及资源
2. 本地环境搭建
3. 本地mock redis,mongo,mysql,safereview
4. test
5. 多版本灰度上线
6. 实现ws的router，重构communication
7. 前后端接口模块，schema生成ts及go


## 构建
docker pull golang:1.22-alpine
docker pull alpine:3.17
docker build  -t kcserver:latest  .



## go编辑慢
go env #查看GOPROXY=https://proxy.golang.org,direct换成国内的其中一个
go env -w  GOPROXY=https://goproxy.cn,direct
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct