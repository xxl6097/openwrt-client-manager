package openwrt

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/openwrt-client-manager/internal/u"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	instance *openWRT
	once     sync.Once
)

type openWRT struct {
	nickMap          map[string]*NickEntry
	clients          map[string]*DHCPLease
	tempClientStatus map[string]bool
	//fsWatcher *fsnotify.Watcher
	fnWatcher func()
}

// GetInstance 返回单例实例
func GetInstance() *openWRT {
	once.Do(func() {
		instance = &openWRT{
			nickMap: make(map[string]*NickEntry),
		}
		instance.init()
		glog.Println("Singleton instance created")
	})
	return instance
}

func (this *openWRT) init() {
	if u.IsMacOs() {
		return
	}
	this.initClients()
	time.Sleep(time.Second)
	go this.initListenSysLog()
	this.initListenFsnotify()
}

func (this *openWRT) initListenSysLog() {
	err := listenSysLog(func(timestamp int64, macAddr string, phy string, status bool) {
		glog.Printf("syslog %s【%s】%v %v\n", timestamp, macAddr, phy, status)
		this.updateClientsBySysLog(timestamp, macAddr, phy, status)
	})
	if err != nil {
		glog.Error(fmt.Errorf("listenSysLog Error:%v", err))
	}
}

// 检测变化并告警
//func (this *openWRT) checkARPDiff(fn func([]string)) {
//	if this.arpList == nil || len(this.arpList) == 0 {
//		return
//	}
//	if fn == nil {
//		return
//	}
//	arpList, err := getArp(brLanString)
//	if err != nil {
//		return
//	}
//	if arpList == nil || len(arpList) == 0 {
//		return
//	}
//	arp1 := strings.Join(arpList, ",")
//	arp2 := strings.Join(this.arpList, ",")
//	if strings.Compare(arp1, arp2) != 0 {
//		this.arpList = arpList
//		fn(arpList)
//	}
//}

func (this *openWRT) listenFsnotify(watcher *fsnotify.Watcher) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			glog.Println("event:", event)
			if event.Has(fsnotify.Write) {
				//filePath := event.Name
				if strings.Compare(strings.ToLower(event.Name), strings.ToLower(dhcpLeasesFilePath)) == 0 {
					this.updateClientsByDHCP()
				}
				if this.fnWatcher != nil {
					this.fnWatcher()
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			glog.Println("error:", err)
		}
	}
}

func (this *openWRT) initListenFsnotify() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		glog.Error(fmt.Errorf("创建监控器失败 %v", err))
	}
	go this.listenFsnotify(watcher)
	// Add a path.
	err = watcher.Add(dhcpLeasesFilePath)
	//err = watcher.Add(arpFilePath)
	if err != nil {
		glog.Error(fmt.Errorf("watcher add err %v", err))
	}
}

func (this *openWRT) Listen(fn func([]*DHCPLease)) {
	this.fnWatcher = func() {
		if fn != nil {
			fn(this.GetClients())
		}
	}
}

func (this *openWRT) GetClients() []*DHCPLease {
	data := make([]*DHCPLease, 0)
	for _, cls := range this.clients {
		data = append(data, cls)
	}
	sort.Slice(data, func(i, j int) bool {
		// 在线状态优先：在线(true) > 离线(false)
		if data[i].Online != data[j].Online {
			return data[i].Online
		}
		return data[i].Hostname < data[j].Hostname
	})
	return data
}
func (this *openWRT) ResetClients() {
	this.initClients()

	if this.fnWatcher != nil {
		this.fnWatcher()
	}
}

func (this *openWRT) initClients() {
	dataMap, err := this.getClientsFromDHCPAndArpAndSysLogAndNick()
	if err != nil {
		glog.Errorf("initClients Error:%v", err)
		time.Sleep(5 * time.Second)
		glog.Error("5 seconds later and try...")
		this.initClients()
	}
	this.clients = dataMap
}

func (this *openWRT) updateClientStatusListBySysLog(macAddr string, timestamp int64, status bool) {
	list := getStatusByMac(macAddr)
	s := Status{
		Timestamp: timestamp,
		Connected: status,
	}
	if list == nil {
		list = make([]*Status, 0)
		statsList, e3 := getStatusFromSysLog()
		if e3 == nil {
			tempList := statsList[macAddr]
			if tempList != nil && len(tempList) > 0 {
				list = append(list, tempList...)
			}

		}
	}
	list = append(list, &s)
	size := len(list)
	if len(list) > MAX_SIZE {
		tempSize := size - MAX_SIZE
		list = list[tempSize:]
	}
	_ = setStatusByMac(macAddr, list)
}

func (this *openWRT) updateClientsBySysLog(timestamp int64, macAddr string, phy string, status bool) {
	if cls, ok := this.clients[macAddr]; ok {
		//cls.StatusList = append(cls.StatusList, &Status{
		//	Timestamp: timestamp,
		//	Connected: status,
		//})
		cls.Online = status
		cls.Phy = phy
		this.tempClientStatus[macAddr] = status
		glog.Infof("updateClientsBySysLog:%v", cls)
		if this.fnWatcher != nil {
			this.fnWatcher()
		}
	}
	this.updateClientStatusListBySysLog(macAddr, timestamp, status)
	//for _, nick := range this.nickMap {
	//	glog.Printf("nick %+v\n", nick)
	//}

}

func (p *openWRT) updateClientsByDHCP() {
	clientArray, err := getClientsByDhcp()
	if err != nil {
		glog.Println(fmt.Errorf("getClientsByDhcp Error:%v", err))
	} else {
		glog.Printf("DHCP更新客户端 %+v\n", len(clientArray))
		for _, client := range clientArray {
			mac := client.MAC
			if status, okk := p.tempClientStatus[mac]; okk {
				client.Online = status
			}
			if p.nickMap != nil {
				if nick, ok := p.nickMap[mac]; ok {
					client.NickName = nick.Name
					if nick.Hostname != "" && nick.Hostname != "*" {
						client.Hostname = nick.Hostname
					}
				}
			}
			if v, ok := p.clients[mac]; ok {
				if client.Hostname != "" && client.Hostname != "*" {
					v.Hostname = client.Hostname
				}
				v.IP = client.IP
				v.StartTime = client.StartTime
				v.NickName = client.NickName
				v.Online = client.Online

			} else {
				p.clients[mac] = client
			}
		}
	}
}

//func (this *openWRT) updateClientsByARP() {
//	clientArray, err := getClientsByArp(brLanString)
//	if err != nil {
//		glog.Println(fmt.Errorf("getClientsByArp Error:%v", err))
//	} else {
//		glog.Printf("ARP更新客户端 %+v\n", len(clientArray))
//		for _, client := range clientArray {
//			mac := client.MAC.String()
//			item := &DHCPLease{
//				IP:     client.IP.String(),
//				MAC:    mac,
//				Online: client.Flags == 2,
//			}
//			if this.nickMap != nil {
//				if nick, ok := this.nickMap[mac]; ok {
//					item.NickName = nick.Name
//					if nick.Hostname != "" && nick.Hostname != "*" {
//						item.Hostname = nick.Hostname
//					}
//				}
//			}
//
//			if v, ok := this.clients[mac]; ok {
//				if item.Hostname != "" && item.Hostname != "*" {
//					v.Hostname = item.Hostname
//				}
//				v.IP = item.IP
//				v.NickName = item.NickName
//				v.Online = item.Online
//			} else {
//				this.clients[mac] = item
//			}
//		}
//	}
//}

func (this *openWRT) initStatusListBySysLog(macAddr string, newList []*Status) {
	list := getStatusByMac(macAddr)
	if list == nil {
		list = newList
	} else {
		element := list[len(list)-1]
		if element != nil {
			for i, n := range newList {
				if n.Timestamp > element.Timestamp {
					list = append(list, newList[i:]...)
				}
			}
		}
	}
	if list == nil {
		return
	}
	size := len(list)
	if len(list) > MAX_SIZE {
		tempSize := size - MAX_SIZE
		list = list[tempSize:]
	}
	_ = setStatusByMac(macAddr, list)
}

func (this *openWRT) getClientsFromDHCPAndArpAndSysLogAndNick() (map[string]*DHCPLease, error) {
	entries, e1 := getClientsByArp(brLanString)
	if e1 == nil {
		leases, e2 := getClientsByDhcp()
		status, e3 := getStatusFromSysLog()
		nicks, e4 := getNickData()
		this.nickMap = nicks
		glog.Errorf("getNickData Error:%v", e4)
		if e4 != nil {
			nicks = map[string]*NickEntry{}
		} else {
			for _, nick := range this.nickMap {
				glog.Debugf("NickData:%+v", nick)
			}
		}
		dataMap := make(map[string]*DHCPLease)
		for _, entry := range entries {
			mac := entry.MAC.String()
			item := &DHCPLease{
				IP:     entry.IP.String(),
				MAC:    mac,
				Online: entry.Flags == 2,
			}
			if e2 == nil {
				if lease, ok := leases[mac]; ok {
					item.StartTime = lease.StartTime
					item.Hostname = lease.Hostname
				}
			}
			if e3 == nil {
				//item.StatusList = status[mac]
				list := status[mac]
				//_ = setStatusByMac(mac, list)
				this.initStatusListBySysLog(mac, list)
			}
			if e4 == nil {
				if nick, ok := nicks[mac]; ok {
					item.NickName = nick.Name
				}
			} else {
				nicks[mac] = &NickEntry{
					StartTime: item.StartTime,
					Hostname:  item.Hostname,
					IP:        item.IP,
					MAC:       mac,
				}
			}
			dataMap[mac] = item
		}
		if e4 != nil {
			err := setNickData(nicks)
			if err != nil {
				glog.Errorf("NickData Save Error:%v", err)
			}
		}
		return dataMap, nil
	}
	return nil, e1
}

func (this *openWRT) Nick(obj *DHCPLease) error {
	if obj == nil {
		return errors.New("DHCPLease obj is nill")
	}
	mac := obj.MAC
	if v, ok := this.nickMap[mac]; ok {
		v.Name = obj.NickName
	}
	if v, ok := this.clients[mac]; ok {
		v.NickName = obj.NickName
	}
	if this.fnWatcher != nil {
		this.fnWatcher()
	}
	return setNickData(this.nickMap)
}

func (this *openWRT) GetDeviceStatusList(mac string) []*Status {
	return getStatusByMac(mac)
}
func (this *openWRT) DeleteStaticIp(mac string) error {
	return deleteStaticIpAddress(mac)
}

func (this *openWRT) GetStaticIps() ([]DHCPHost, error) {
	return GetUCIOutput()
}

func (this *openWRT) SetStaticIp(mac, ip, name string) error {
	return setStaticIpAddress(mac, ip, name)
}
