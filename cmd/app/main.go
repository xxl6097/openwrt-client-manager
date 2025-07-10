package main

import (
	"fmt"
	"github.com/xxl6097/go-http/pkg/httpserver"
	assets "github.com/xxl6097/openwrt-client-manager/assets/openwrt"
	"github.com/xxl6097/openwrt-client-manager/internal"
	"github.com/xxl6097/openwrt-client-manager/internal/u"
	"github.com/xxl6097/openwrt-client-manager/pkg"
)

func init() {
	if u.IsMacOs() {
		pkg.BinName = "openwrt-client-manager_v0.0.20_darwin_arm64"
	}
}
func main() {
	fmt.Println("Hello World")
	httpserver.New().
		CORSMethodMiddleware().
		BasicAuth("admin", "admin").
		AddRoute(internal.NewRoute(internal.NewApi(nil))).
		AddRoute(assets.NewRoute()).
		Done(8080)
}
