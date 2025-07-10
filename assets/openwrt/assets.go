package assets

import (
	"e.coding.net/clife-devops/devp/go-http/pkg/httpserver"
	"e.coding.net/clife-devops/devp/go-http/pkg/ihttpserver"
	"e.coding.net/clife-devops/devp/go-http/pkg/util"
	"embed"
	"github.com/gorilla/mux"
	"github.com/xxl6097/glog/glog"
	"io/fs"
	"net/http"
)

//go:embed static/*
var StaticFS embed.FS
var FileSystem http.FileSystem

func init() {
	subFs, err := fs.Sub(StaticFS, "static")
	if err != nil {
		glog.Fatal("静态资源加载失败", err)
	}
	FileSystem = http.FS(subFs)
}

type StaticRoute struct {
}

func (s StaticRoute) Setup(router *mux.Router) {
	httpserver.RouterUtil.AddNoAuthPrefix("/")
	httpserver.RouterUtil.AddNoAuthPrefix("static")

	router.Handle("/favicon.ico", http.FileServer(FileSystem)).Methods(http.MethodGet, http.MethodOptions)
	router.PathPrefix("/").Handler(util.MakeHTTPGzipHandler(http.StripPrefix("/", http.FileServer(FileSystem)))).Methods(http.MethodGet, http.MethodOptions)
}

func NewRoute() ihttpserver.IRoute {
	opt := &StaticRoute{}
	return opt
}
