package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/docker/client"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: injecto <image> <container>")
		os.Exit(1)
	}

	image := processImage(os.Args[1])
	container := os.Args[2]

	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	if err := save(cli, dir, image); err != nil {
		panic(err)
	}
	if err := copy(cli, dir, container); err != nil {
		panic(err)
	}
}

func processImage(s string) string {
	parts := strings.Split(s, ":")
	if len(parts) == 1 {
		return strings.Join([]string{s, "latest"}, ":")
	}
	return s
}
