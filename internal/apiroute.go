package internal

import (
	"e.coding.net/clife-devops/devp/go-http/pkg/httpserver"
	"e.coding.net/clife-devops/devp/go-http/pkg/ihttpserver"
	"github.com/gorilla/mux"
	"net/http"
)

type ApiRoute struct {
	restApi *Api
}

func NewRoute(ctl *Api) ihttpserver.IRoute {
	opt := &ApiRoute{
		restApi: ctl,
	}
	return opt
}

func (this *ApiRoute) Setup(router *mux.Router) {
	httpserver.RouterUtil.AddHandleFunc(router, ihttpserver.ApiModel{
		Method: http.MethodGet,
		Path:   "/api/get/clients",
		Fun:    this.restApi.GetClients,
		NoAuth: false,
	})
	httpserver.RouterUtil.AddHandleFunc(router, ihttpserver.ApiModel{
		Method: http.MethodGet,
		Path:   "/get/status",
		Fun:    this.restApi.GetStatus,
		NoAuth: false,
	})
	httpserver.RouterUtil.AddHandleFunc(router, ihttpserver.ApiModel{
		Method: http.MethodDelete,
		Path:   "/api/clear",
		Fun:    this.restApi.Clear,
		NoAuth: false,
	})
	httpserver.RouterUtil.AddHandleFunc(router, ihttpserver.ApiModel{
		Method: http.MethodPost,
		Path:   "/api/nick/set",
		Fun:    this.restApi.SetNick,
		NoAuth: false,
	})

	router.Handle("/api/client/sse", this.restApi.GetSSE())
	//subRouter.Handle("/api/client/sse", this.sseApi)
	//httpserver.RouterUtil.AddHandleFunc(router, ihttpserver.ApiModel{
	//	Method: http.MethodPost,
	//	Path:   "/frp",
	//	Fun:    this.controller.Frp,
	//	NoAuth: false,
	//})
}
