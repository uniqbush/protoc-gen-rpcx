package utils

import (
	"fmt"
	"io/ioutil"
	"os"
)

func GenerateMod(path string, service string) {
	file := path + "/go.mod"
	exist := IsExist(file)
	if !exist {
		content := fmt.Sprintf(`module %s

go 1.16

require (
	github.com/smallnest/rpcx v1.6.4
)
`, service)
		err := ioutil.WriteFile(file, []byte(content), 0644)
		if err != nil {
			panic(err)
		}
	}
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}

		if os.IsNotExist(err) {
			return false
		}

		fmt.Println(err)
		return false
	}

	return true
}

func WriteFile(filename string, content string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.WriteString(content); err != nil {
		panic(err)
	}
}
