package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"i-am-the-emperor/backend/server"
)

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	app := server.New()
	fmt.Printf("我是皇帝 running at http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, app))
}
