# taskchain
一个基于区块链的任务管理应用。

## 安装
```sh
$ git clone https://github.com/zhanet/taskchain

$ cd taskchain
```

### 安装 dep 和依赖包
```sh
$ go get -u github.com/golang/dep/cmd/dep 

对于 Mac，可用 brew install dep

$ dep ensure
```

## 运行
```sh
$ echo PORT=9000 > .env

$ go run main.go
```

## API
当前只有两个API：
- GET获取区块链数据
- POST提交区块数据

## 测试

可以使用 [Postman](https://www.getpostman.com/apps) 或 cURL，下面以 cURL 为例。

新开一个shell窗口或Tab，输入以下命令：

```sh
$ curl -X GET \
  http://localhost:9000/ \
  -H 'cache-control: no-cache'
```
```sh
$ curl -X POST \
  http://localhost:9000/ \
  -H 'cache-control: no-cache' \
  -H 'content-type: application/json' \
  -d '{"title":"t01", "description":"test block01"}'
```
