package assets

import (
	"e.coding.net/clife-devops/devp/go-http/pkg/httpserver"
	"e.coding.net/clife-devops/devp/go-http/pkg/ihttpserver"
	"e.coding.net/clife-devops/devp/go-http/pkg/util"
	"github.com/gorilla/mux"
	"net/http"
)

func init() {
	Load("")
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
