# grpc-example

> go grpc example

- grpc client load balance example

## 新手教程

### 1. 安装 protobuf

```shell
$ brew install protobuf
# 检查是否安装成功
$ protoc --version
```

### 2. 安装 go 语言代码生成工具

go install 需要 go1.16 版本

```shell
# 官方工具
$ go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
$ go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
# 使用 gogo/protobuf 优化代码生成
$ go install github.com/gogo/protobuf/protoc-gen-gogofast@latest
```

[./scripts/install.sh](./scripts/install.sh)

### 3. 生成代码

介绍官方生成器和 gogo/protobuf 两种生成器, 新手可以忽略掉 gogo/protobuf

#### 使用官方版本代码生成器

```shell
#!/bin/bash -x
set -e

OUT_DIR=./pb

rm -rf $OUT_DIR/*.{go,json}

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    pb/origin-hello.proto

```

[./scripts/origin-gen.sh](./scripts/origin-gen.sh) 对应 proto 文件 [./pb/origin-hello.proto](./pb/origin-hello.proto)

#### 使用 gogo/protobuf 生成

```shell
#!/bin/bash -x
set -e

OUT_DIR=./pb

GOGOPROTO_ROOT="$(GO111MODULE=on go list -m -f '{{ .Dir }}' -m github.com/gogo/protobuf)"
GOGOPROTO_PATH="${GOGOPROTO_ROOT}:${GOGOPROTO_ROOT}/protobuf"

rm -rf $OUT_DIR/*.{go,json}

protoc --gogofast_out=plugins=grpc:. \
    -I=. \
    -I="${GOGOPROTO_PATH}" \
    pb/hello.proto

```

[./scripts/gen.sh](./scripts/gen.sh) 对应 proto 文件 [./pb/hello.proto](./pb/hello.proto)

### 4. 编写代码

服务端实现生成接口 `pb.HelloServer` 接口, 客户端直接使用 `pb.NewHelloClient` 初始化 client 调用接口.

## License

MIT &copy; zcong1993
