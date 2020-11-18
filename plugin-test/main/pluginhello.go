package main

import "fmt"

//go build --buildmode=plugin -o pluginhello.so pluginhello.go
//编译为插件，只能在linux里面执行
func Hello() {
	fmt.Print("Hello world From Plugin!")
}

func Invoke(n int) string {
	return fmt.Sprintf("调用我传了一个参数n:%d", n)
}
