package openwrt

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/xxl6097/glog/glog"
	"github.com/xxl6097/go-service/pkg/ukey"
	"github.com/xxl6097/openwrt-client-manager/internal/u"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	nickFilePath          = "/usr/local/nick"
	arpFilePath           = "/proc/net/arp"
	dhcpLeasesFilePath    = "/tmp/dhcp.leases"
	dhcpCfgFilePath       = "/etc/config/dhcp"
	brLanString           = "br-lan"
	apStaDisConnectString = "AP-STA-DISCONNECTED"
	apStaConnectString    = "AP-STA-CONNECTED"
	statusDir             = "/usr/local/openwrt/status"
	MAX_SIZE              = 1000
)

type Status struct {
	Timestamp int64 `json:"timestamp"`
	Connected bool  `json:"connected"`
	//MAC       string `json:"mac"`
}

type DHCPLease struct {
	IP        string `json:"ip"`  //DHCP 服务器分配给客户端的 IP
	MAC       string `json:"mac"` //设备的物理地址，格式为 xx:xx:xx:xx:xx:xx
	Phy       string `json:"phy"`
	Hostname  string `json:"hostname"` //客户端上报的主机名（可能为空或 *）
	NickName  string `json:"nickName"` //
	StartTime int64  `json:"starTime"` //租约失效的精确时间（秒级精度）
	Online    bool   `json:"online"`
	//StatusList []*Status `json:"statusList"`
}
type ARPEntry struct {
	IP        net.IP           //设备的 IPv4 地址
	HWType    uint8            // 硬件类型（通常为 0x1，表示以太网）。
	Flags     uint8            // ARP 表项状态标志：0x0：无效（离线）;0x2：有效（在线），表示设备可达。
	MAC       net.HardwareAddr //设备的 MAC 地址
	Mask      string           // 子网掩码（通常为 *，表示未使用）
	Interface string           // 关联的网络接口（如 br-lan、eth0）
}
type NickEntry struct {
	Name      string `json:"name"`
	StartTime int64  `json:"starTime"`
	MAC       string `json:"mac"`
	IP        string `json:"ip"`
	Hostname  string `json:"hostname"`
}

func getDataFromSysLog(pattern string, args ...string) (map[string][]DHCPLease, error) {
	dataMap := make(map[string][]DHCPLease)
	// 2. 编译正则表达式（匹配连接/断开事件）
	//pattern := `AP-STA-(CONNECTED|DISCONNECTED)`
	re := regexp.MustCompile(pattern)
	return dataMap, command(func(data string) {
		if re.MatchString(data) {
			//fmt.Println("[事件] ", data) // 输出匹配行
			macAddr := parseMacAddr(data)
			mac, err := net.ParseMAC(macAddr)
			if err == nil {
				t, e := parseTimer(data)
				if e == nil {
					var status bool
					if strings.Contains(data, apStaConnectString) {
						status = true
					} else if strings.Contains(data, apStaDisConnectString) {
						status = false
					}
					element := DHCPLease{
						MAC:       mac.String(),
						StartTime: t.Unix(),
						Online:    status,
					}
					v, ok := dataMap[element.MAC]
					if ok {
						v = append(v, element)
					} else {
						v = []DHCPLease{element}
					}
					dataMap[element.MAC] = v
				}
			}
		}

	}, "logread", args...)
}

func listenSysLog(fn func(int64, string, string, bool)) error {
	//args := []string{"-f", "|", "grep", "hostapd.*"}
	pattern := `hostapd.*`
	re := regexp.MustCompile(pattern)
	return command(func(s string) {
		if re.MatchString(s) {
			timestamp, macAddr, phy, status := parseSysLog(s)
			//if fn != nil {
			//	fn(timestamp, macAddr, status == 0)
			//}
			if status == 0 {
				if fn != nil {
					fn(timestamp, macAddr, phy, false)
				}
			} else if status == 1 {
				if fn != nil {
					fn(timestamp, macAddr, phy, true)
				}
			} else {
				//fmt.Printf("未知类型 %s\n", s)
			}
		}

	}, "logread", "-f")
}

func getStatusFromSysLog() (map[string][]*Status, error) {
	pattern := `AP-STA-(CONNECTED|DISCONNECTED)`
	data, err := getDataFromSysLog(pattern)
	//fmt.Printf("GetDisconnectFromSysLog %+v\n", data)
	if err == nil && len(data) > 0 {
		times := make(map[string][]*Status)
		for k, v := range data {
			value := make([]*Status, 0)
			for _, entry := range v {
				value = append(value, &Status{
					//MAC:       entry.MAC,
					Timestamp: entry.StartTime,
					Connected: entry.Online,
				})
			}
			times[k] = value
		}
		return times, err
	}
	return nil, err
}

func parseARPLine(line string) (*ARPEntry, error) {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return nil, fmt.Errorf("invalid ARP line: expected 6 fields, got %d", len(fields))
	}

	// 解析 IP 地址
	ip := net.ParseIP(fields[0])
	if ip == nil {
		return nil, fmt.Errorf("invalid IP: %s", fields[0])
	}

	// 解析十六进制数值（HWType 和 Flags）
	hwType, _ := strconv.ParseUint(strings.TrimPrefix(fields[1], "0x"), 16, 8)
	flags, _ := strconv.ParseUint(strings.TrimPrefix(fields[2], "0x"), 16, 8)

	// 解析 MAC 地址
	mac, err := net.ParseMAC(fields[3])
	if err != nil {
		return nil, fmt.Errorf("invalid MAC: %v", err)
	}
	if mac.String() == "00:00:00:00:00:00" {
		return nil, fmt.Errorf("error MAC")
	}

	return &ARPEntry{
		IP:        ip,
		HWType:    uint8(hwType),
		Flags:     uint8(flags),
		MAC:       mac,
		Mask:      fields[4],
		Interface: fields[5],
	}, nil
}

func getLeaseTime() time.Duration {
	data, err := os.ReadFile(dhcpCfgFilePath)
	if err != nil {
		fmt.Println("读取文件失败:", err)
		return time.Duration(3600 * 12)
	}

	// 正则匹配leasetime选项
	re := regexp.MustCompile(`option leasetime ['"]([\dhms]+)['"]`)
	match := re.FindStringSubmatch(string(data))
	if len(match) < 2 {
		fmt.Println("未找到leasetime配置")
		return time.Duration(3600 * 12)
	}

	// 解析时间字符串（如"12h"）
	leaseStr := match[1]
	leaseDuration, err := time.ParseDuration(leaseStr)
	if err != nil {
		fmt.Println("解析时间失败:", err)
		return time.Duration(3600 * 12)
	}
	//fmt.Printf("DHCP租约时间: %v（%d秒）\n", leaseStr, int(leaseDuration.Seconds()))
	return leaseDuration
}

func parseTime(logLine string) int64 {
	//re := regexp.MustCompile(`^(\w+\s+\w+\s+\d+\s+\d+:\d+:\d+\s+\d+)`)
	//matches := re.FindStringSubmatch(logLine)
	//if len(matches) > 1 {
	//	timeStr := matches[1]
	//	t, err := autoParse(timeStr)
	//	if err == nil {
	//		return t.Format(time.DateTime)
	//	}
	//}
	//return ""
	t, err := parseTimer(logLine)
	if err == nil {
		return t.Unix()
	}
	return 0
}

func parseTime1(logLine string) string {
	t, err := parseTimer(logLine)
	if err == nil {
		return t.Format(time.DateTime)
	}
	return ""
}

func parseMacAddr(logLine string) string {
	// 1. 定义MAC地址正则表达式（兼容冒号/短横线分隔）
	pattern := `(?:[0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}`
	re := regexp.MustCompile(pattern)
	// 2. 提取所有匹配的MAC地址
	macAddresses := re.FindAllString(logLine, -1)
	// 3. 输出结果
	if len(macAddresses) > 0 {
		return macAddresses[0]
	}
	return ""
}
func parseSysLog(data string) (int64, string, string, int) {
	phy := parsePhy(data)
	timestamp := parseTime(data)
	macAddr := parseMacAddr(data)
	// 1. 检查字符串是否包含目标字段
	if strings.Contains(data, apStaDisConnectString) { //AP-STA-DISCONNECTED
		//fmt.Printf("%s 设备【%s】连上了", timestamp, macAddr)
		return timestamp, macAddr, phy, 0
	} else if strings.Contains(data, apStaConnectString) { //AP-STA-CONNECTED
		//fmt.Printf("%s 设备【%s】断开了", timestamp, macAddr)
		return timestamp, macAddr, phy, 1
	}
	return timestamp, macAddr, phy, -1
}

func parseLeaseLine(line string, leasetime time.Duration) (DHCPLease, error) {
	// 示例行: 1693837890 00:11:22:33:44:55 192.168.1.100 hostname-1
	fields := strings.Fields(line)
	if len(fields) < 4 { // 至少包含时间戳、MAC、IP、主机名
		return DHCPLease{}, fmt.Errorf("字段不足")
	}

	//fmt.Println("fields", fields)
	// 解析时间戳（Unix时间）
	startSec, _ := strconv.ParseInt(fields[0], 10, 64)
	beijingLoc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		// 备选方案：手动创建东八区时区
		//fmt.Println(err, "备选方案：手动创建东八区时区")
		//beijingLoc = time.FixedZone("CST", 8*60*60) // UTC+8
		beijingLoc = time.FixedZone("UTC+8", 8*60*60)
	}
	utcTime := time.Unix(startSec, 0) //// 解析为 UTC 时间
	startTime := utcTime.In(beijingLoc).Add(-leasetime)
	//startTime := time.Unix(startSec, 0) //// 解析为 UTC 时间

	// 解析MAC地址
	mac, err := net.ParseMAC(fields[1])
	if err != nil {
		return DHCPLease{}, fmt.Errorf("MAC格式错误: %v", err)
	}

	// 解析IP地址
	ip := net.ParseIP(fields[2])
	if ip == nil {
		return DHCPLease{}, fmt.Errorf("IP格式错误")
	}

	// 主机名（可能包含空格，合并剩余字段）
	//hostname := strings.Join(fields[3:], " ")
	hostname := fields[3]

	return DHCPLease{
		IP:        ip.String(),
		MAC:       mac.String(),
		Hostname:  hostname,
		StartTime: startTime.Unix(),
		//IsActive:  time.Now().Before(startTime.Add(time.Second * time.Duration(leaseDuration))),
	}, nil
}

func getNickData() (map[string]*NickEntry, error) {
	data, err := os.ReadFile(nickFilePath)
	if err != nil {
		return nil, err
	}
	dataMap := map[string]*NickEntry{}
	err = json.Unmarshal(data, &dataMap)
	if err != nil {
		return nil, err
	}
	return dataMap, nil
}

func setNickData(dataMap map[string]*NickEntry) error {
	if dataMap == nil || len(dataMap) == 0 {
		return nil
	}
	content, err := json.Marshal(dataMap)
	if err != nil {
		return err
	}
	file, err := os.Create(nickFilePath) // 文件不存在则创建，存在则截断
	if err != nil {
		return err
	}
	defer file.Close()
	// 写入内容
	_, err = file.Write(content)
	return err
}

func getArp(deviceInterfaceName string) ([]string, error) {
	data, err := os.ReadFile(arpFilePath)
	if err != nil {
		return nil, err
	}
	arpText := make([]string, 0)
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // 跳过标题行和空行
		}
		if !strings.HasSuffix(line, deviceInterfaceName) {
			continue // 根据Device过滤
		}
		arpText = append(arpText, strings.TrimSpace(line))
	}
	return arpText, nil
}

func getClientsByArp(deviceInterfaceName string) (map[string]*ARPEntry, error) {
	data, err := os.ReadFile(arpFilePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	//entries := make([]*ARPEntry, 0)
	entries := make(map[string]*ARPEntry)
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // 跳过标题行和空行
		}
		if !strings.HasSuffix(line, deviceInterfaceName) {
			continue // 根据Device过滤
		}
		entry, e := parseARPLine(line)
		if e != nil {
			//return nil, err
			glog.Error("parseARPLine error", e, line)
			continue
		}
		entries[entry.MAC.String()] = entry
	}
	return entries, nil
}

func parsePhy(logLine string) string {
	re := regexp.MustCompile(`hostapd:\s+(phy[\w-]+?):`)
	// 提取匹配结果
	matches := re.FindStringSubmatch(logLine)
	if len(matches) < 2 {
		return ""
	}
	phyField := matches[1] // 捕获组索引为1
	return phyField
}

func parseArpLines(lines []string) (map[string]*ARPEntry, error) {
	if lines == nil || len(lines) == 0 {
		return nil, nil
	}
	entries := make(map[string]*ARPEntry)
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // 跳过标题行和空行
		}
		entry, err := parseARPLine(line)
		if err != nil {
			return nil, err
		}
		entries[entry.MAC.String()] = entry
	}
	return entries, nil
}
func parseDHCPLeases(filePath string) (map[string]*DHCPLease, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	leaseTime := getLeaseTime()
	entries := make(map[string]*DHCPLease)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // 跳过注释和空行
		}
		lease, err := parseLeaseLine(line, leaseTime)
		if err != nil {
			log.Printf("解析失败: %v | 行: %s", err, line)
			continue
		}
		entries[lease.MAC] = &lease
	}
	return entries, nil
}

func getClientsByDhcp() (map[string]*DHCPLease, error) {
	return parseDHCPLeases(dhcpLeasesFilePath)
}

func parseTimer(logLine string) (*time.Time, error) {
	re := regexp.MustCompile(`^(\w+\s+\w+\s+\d+\s+\d+:\d+:\d+\s+\d+)`)
	matches := re.FindStringSubmatch(logLine)
	if len(matches) > 1 {
		timeStr := matches[1]
		t, err := autoParse(timeStr)
		return &t, err
	}
	return nil, nil
}

func autoParse(timeStr string) (time.Time, error) {
	var layouts = []string{
		time.Layout,
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
		time.Kitchen,
		time.Stamp,
		time.StampMilli,
		time.StampMicro,
		time.StampNano,
		time.DateTime,
		time.DateOnly,
		time.TimeOnly,
	}
	for _, layout := range layouts {
		t, err := time.Parse(layout, timeStr)
		if err == nil {
			return t, nil // 解析成功
		}
	}
	return time.Time{}, fmt.Errorf("无法识别的格式")
}

func command(fu func(string), name string, arg ...string) error {
	fmt.Println(name, arg)
	// 创建ubus命令对象
	cmd := exec.Command(name, arg...)

	// 创建标准输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("创建管道失败: %v\n", err)
		return err
	}

	// 启动命令
	if e := cmd.Start(); e != nil {
		fmt.Printf("启动命令失败: %v\n", e)
		return err
	}
	defer cmd.Process.Kill() // 确保退出时终止进程

	// 实时读取输出流
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		rawEvent := scanner.Text()
		//fmt.Printf("原始事件: %s\n", rawEvent)
		fu(rawEvent)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("读取错误: %v\n", err)
		return err
	}
	return cmd.Wait() // 等待命令退出
}

func getStatusByMac(mac string) []*Status {
	if mac == "" {
		return nil
	}
	_ = u.CheckDirector(statusDir)
	tempFilePath := filepath.Join(statusDir, mac)
	byteArray, err := os.ReadFile(tempFilePath)
	if err != nil {
		return nil
	}
	var cfg []*Status
	err = ukey.GobToStruct(byteArray, &cfg)
	if err != nil {
		return nil
	}
	return cfg
}

func setStatusByMac(mac string, statusList []*Status) error {
	if mac == "" {
		return nil
	}
	if statusList == nil || len(statusList) == 0 {
		return nil
	}

	content, err := ukey.StructToGob(statusList)
	if err != nil {
		return err
	}

	_ = u.CheckDirector(statusDir)
	tempFilePath := filepath.Join(statusDir, mac)
	file, err := os.Create(tempFilePath) // 文件不存在则创建，存在则截断
	if err != nil {
		return err
	}
	defer file.Close()
	// 写入内容
	_, err = file.Write(content)
	return err
}
