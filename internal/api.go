package internal

import (
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/openwrt-client-manager/internal/iface"
	"github.com/xxl6097/openwrt-client-manager/internal/openwrt"
	"github.com/xxl6097/openwrt-client-manager/internal/sse"
	"github.com/xxl6097/openwrt-client-manager/internal/u"
	"net/http"
)

type Api struct {
	sseApi iface.ISSE
}

func NewApi() *Api {
	sseApi := sse.NewServer()
	sseApi.Start()
	a := &Api{
		sseApi: sseApi,
	}
	openwrt.GetInstance().Listen(a.listen)
	return a
}

func (this *Api) listen(list []*openwrt.DHCPLease) {
	if len(list) >= 0 && this.sseApi != nil {
		eve := iface.SSEEvent{
			Event:   "update",
			Payload: list,
		}
		this.sseApi.Broadcast(eve)
	}
}

func (this *Api) GetClients(w http.ResponseWriter, r *http.Request) {
	//req := utils.GetReqMapData(w, r)
	//glog.Warn(req)
	//glog.Warn("getClients---->", r)
	//cls, err := getClients()
	//
	//if err != nil {
	//	glog.Error("getClients err:", err)
	//	u.Respond(w, u.Error(-1, err.Error()))
	//} else {
	//
	//}
	u.Respond(w, u.SucessWithObject(openwrt.GetInstance().GetClients()))
}

func (this *Api) GetStatus(w http.ResponseWriter, r *http.Request) {
	//req := utils.GetReqMapData(w, r)
	//glog.Warn(req)
	glog.Warn("GetStatus---->", r)
	//status, err := getStatusFromSysLog()
	//if err != nil {
	//	glog.Error("getClients err:", err)
	//	u.Respond(w, u.Error(-1, err.Error()))
	//} else {
	//	u.Respond(w, u.SucessWithObject(status))
	//}
}

func (this *Api) Clear(w http.ResponseWriter, r *http.Request) {
	//req := utils.GetReqMapData(w, r)
	//glog.Warn(req)
	glog.Warn("Clear---->", r.URL)
	err := u.ClearTemp()
	if err != nil {
		glog.Error("Clear err:", err)
		u.Respond(w, u.Error(-1, err.Error()))
	} else {
		u.OKK(w)
	}
}

func (this *Api) SetNick(w http.ResponseWriter, r *http.Request) {
	body, err := u.GetDataByJson[openwrt.DHCPLease](r)
	if err != nil {
		glog.Error(err)
		u.Respond(w, u.Error(-1, err.Error()))
		return
	}
	err = openwrt.GetInstance().Nick(body)
	if err != nil {
		glog.Error(err)
		u.Respond(w, u.Error(-2, err.Error()))
		return
	}
	u.OKK(w)
}

func (this *Api) GetSSE() iface.ISSE {
	return this.sseApi
}
