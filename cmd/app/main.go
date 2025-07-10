package main

import (
	"e.coding.net/clife-devops/devp/go-http/pkg/httpserver"
	"fmt"
	assets "github.com/xxl6097/openwrt-client-manager/assets/openwrt"
	"github.com/xxl6097/openwrt-client-manager/internal"
)

func main() {
	fmt.Println("Hello World")
	httpserver.New().
		CORSMethodMiddleware().
		AddRoute(internal.NewRoute(internal.NewApi())).
		AddRoute(assets.NewRoute()).
		Done(8080)
}
