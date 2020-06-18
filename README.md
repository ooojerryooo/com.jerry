#1、protocal buffer安装
    https://github.com/google/protobuf/releases下载安装包
    解压后看到protoc.exe 我这里是windows
    最后设置环境变量即可
    
#2、安装 golang protobuf
    go get -u github.com/golang/protobuf/proto 
    go get -u github.com/golang/protobuf/protoc-gen-go
    
#3、安装 gRPC-go
    go get google.golang.org/grpc
    
#4、编译proto文件
    protoc --go_out=plugins=grpc:. helloworld.proto