package main

import (
	"fmt"
	"os"
	"plugin"
)

func main() {
	plug, err := plugin.Open("pluginhello.so")
	if err != nil {
		fmt.Println("error open plugin: ", err)
		os.Exit(-1)
	}
	s, err := plug.Lookup("Hello")
	if err != nil {
		fmt.Println("error lookup Hello: ", err)
		os.Exit(-1)
	}
	if hello, ok := s.(func()); ok {
		hello()
	}

	invoke, err := plug.Lookup("Invoke")
	if err != nil {
		fmt.Println("error lookup Hello: ", err)
		os.Exit(-1)
	}

	newFunc, ok := invoke.(func(n int) string)
	if !ok {
		fmt.Errorf("Plugin does not implement 'Invoke' function!")
	}

	result := newFunc(7)
	fmt.Printf(result)
}
