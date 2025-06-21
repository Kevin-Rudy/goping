package main

import (
	"log"
	"os"
)

func main() {
	// 创建CLI应用
	app := createCliApp()

	// 运行应用
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
