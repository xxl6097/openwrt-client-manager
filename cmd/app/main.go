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
		BasicAuth("admin", "admin").
		AddRoute(internal.NewRoute(internal.NewApi(nil))).
		AddRoute(assets.NewRoute()).
		Done(8080)
}
