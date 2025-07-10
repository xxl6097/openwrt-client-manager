package service

import (
	"e.coding.net/clife-devops/devp/go-http/pkg/httpserver"
	"fmt"
	"github.com/kardianos/service"
	"github.com/xxl6097/glog/glog"
	_ "github.com/xxl6097/go-service/assets/buffer"
	"github.com/xxl6097/go-service/pkg/gs/igs"
	"github.com/xxl6097/go-service/pkg/ukey"
	"github.com/xxl6097/go-service/pkg/utils"
	assets "github.com/xxl6097/openwrt-client-manager/assets/openwrt"
	"github.com/xxl6097/openwrt-client-manager/internal"
	"github.com/xxl6097/openwrt-client-manager/pkg"
	"os"
)

type Service struct {
	timestamp string
	gs        igs.Service
}

func (this *Service) OnFinish() {
}

type Config struct {
	ServerPort int `json:"serverPort"`
}

func load() (*Config, error) {
	defer glog.Flush()
	byteArray, err := ukey.Load()
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = ukey.GobToStruct(byteArray, &cfg)
	//err = json.Unmarshal(byteArray, &cfg)
	if err != nil {
		glog.Println("ClientConfig解析错误", err)
		return nil, err
	}
	pkg.Version()
	return &cfg, nil
}

func (this *Service) OnConfig() *service.Config {
	cfg := service.Config{
		Name:        pkg.AppName,
		DisplayName: pkg.DisplayName,
		Description: pkg.Description,
	}
	return &cfg
}

func (this *Service) OnVersion() string {
	pkg.Version()
	cfg, err := load()
	if err == nil {
		glog.Debugf("cfg:%+v", cfg)
	}
	return pkg.AppVersion
}

func (this *Service) OnRun(service igs.Service) error {
	this.gs = service
	cfg, err := load()
	if err != nil {
		return err
	}
	glog.Debug("程序运行", os.Args)
	httpserver.New().
		CORSMethodMiddleware().
		AddRoute(internal.NewRoute(internal.NewApi())).
		AddRoute(assets.NewRoute()).
		Done(cfg.ServerPort)
	return nil
}

func (this *Service) GetAny(s2 string) []byte {
	return this.menu()
}

func (this *Service) menu() []byte {
	port := utils.InputIntDefault(fmt.Sprintf("输入服务端口(%d)：", 7070), 7070)
	cfg := &Config{ServerPort: port}
	bb, e := ukey.StructToGob(cfg)
	if e != nil {
		return nil
	}
	return bb
}
