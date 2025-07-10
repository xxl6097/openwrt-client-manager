package main

import (
	"fmt"
	"github.com/xxl6097/go-http/pkg/httpserver"
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
