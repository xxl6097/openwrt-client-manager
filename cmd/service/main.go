package main

import (
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/pkg/gs"
	"github.com/xxl6097/openwrt-client-manager/cmd/service/service"
	"github.com/xxl6097/openwrt-client-manager/pkg"
)

func main() {
	pkg.Version()
	s := service.Service{}
	err := gs.Run(&s)
	glog.Debug("程序结束", err)
}
