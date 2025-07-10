package main

import (
	"bufio"
	"e.coding.net/clife-devops/devp/go-http/pkg/httpserver"
	"encoding/json"
	"fmt"
	assets "github.com/xxl6097/openwrt-client-manager/assets/openwrt"
	"github.com/xxl6097/openwrt-client-manager/internal"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func command(fu func(string), name string, arg ...string) {
	// 创建ubus命令对象
	cmd := exec.Command(name, arg...)

	// 创建标准输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("创建管道失败: %v\n", err)
		os.Exit(1)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		fmt.Printf("启动命令失败: %v\n", err)
		return
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
	}
	cmd.Wait() // 等待命令退出
}

func shell() {
	// 创建ubus命令对象
	//ubus listen hostapd.phy1-ap0
	cmd := exec.Command("ubus", "subscribe", "hostapd.phy1-ap0")

	// 创建标准输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("创建管道失败: %v\n", err)
		os.Exit(1)
	}

	// 启动命令
	if err := cmd.Start(); err != nil {
		fmt.Printf("启动命令失败: %v\n", err)
		return
	}
	defer cmd.Process.Kill() // 确保退出时终止进程

	// 实时读取输出流
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		rawEvent := scanner.Text()
		// 解析JSON事件
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(rawEvent), &event); err != nil {
			fmt.Printf("JSON解析失败: %v\n", err)
			continue
		}

		// 提取事件类型和MAC地址
		if data, ok := event["data"]; ok {
			eventData := data.(map[string]interface{})
			eventType := eventData["event"]
			macAddr := eventData["mac"]
			fmt.Printf("事件类型: %s, MAC: %s\n", eventType, macAddr)
		}

		fmt.Printf("事件类型: %v\n", event)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("读取错误: %v\n", err)
	}
	cmd.Wait() // 等待命令退出
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

func parseTime(logLine string) string {
	re := regexp.MustCompile(`^(\w+\s+\w+\s+\d+\s+\d+:\d+:\d+\s+\d+)`)
	matches := re.FindStringSubmatch(logLine)
	if len(matches) > 1 {
		timeStr := matches[1]
		t, err := autoParse(timeStr)
		if err == nil {
			return t.Format(time.DateTime)
		}
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

func parseData(data string) {
	timestamp := parseTime(data)
	macAddr := parseMacAddr(data)
	// 1. 检查字符串是否包含目标字段
	if strings.Contains(data, "AP-STA-DISCONNECTED") {
		fmt.Printf("%s 设备【%s】连上了", timestamp, macAddr)
	} else if strings.Contains(data, "AP-STA-CONNECTED") {
		fmt.Printf("%s 设备【%s】断开了", timestamp, macAddr)
	}
}

func test() {
	logLine := "Fri Jul  4 16:55:51 2025 daemon.notice hostapd: phy0-ap0: AP-STA-DISCONNECTED 28:59:23:cf:8e:3b"
	parseData(logLine)
}
func testParseDHCP() {
	//input := "1751665780 28:59:23:cf:8e:3b 192.168.23.110 Xiaomi-15 01:28:59:23:cf:8e:3b"
	//res := internal.ParseDHCP(input)
	//fmt.Printf("%+v\n", res)
}

func main() {
	fmt.Println("Hello World")
	//leases, err := internal.GetClientsByDhcp()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//for _, lease := range leases {
	//	fmt.Printf("IP: %-15s MAC: %-17s 主机名: %-20s 时间: %v\n",
	//		lease.IP, lease.MAC, lease.Hostname, lease.StartTime)
	//}

	//data := "IP address       HW type     Flags       HW address            Mask     Device\n192.168.0.2      0x1         0x2         06:e4:4a:83:43:3c     *        br-lan\n10.6.50.52       0x1         0x2         50:64:2b:14:f3:ae     *        wan\n10.6.50.110      0x1         0x2         08:5d:8a:8d:be:d3     *        wan\n10.6.50.254      0x1         0x2         1c:20:db:8d:96:76     *        wan\n192.168.0.32     0x1         0x0         00:00:00:00:00:00     *        br-lan\n10.6.50.191      0x1         0x2         cc:96:e5:00:7a:a6     *        wan\n10.6.50.156      0x1         0x2         c0:25:a5:cb:a2:12     *        wan\n10.6.50.100      0x1         0x2         48:4d:7e:b9:3f:12     *        wan\n192.168.0.3      0x1         0x2         8c:ec:4b:58:81:09     *        br-lan\n192.168.0.110    0x1         0x2         28:59:23:cf:8e:3b     *        br-lan\n192.168.0.138    0x1         0x0         62:ae:10:87:f4:47     *        br-lan\n192.168.0.221    0x1         0x2         0e:24:e8:3c:5d:c4     *        br-lan"
	//entries, err := internal.GetClients()
	//if err != nil {
	//	log.Fatalf("Failed to parse ARP table: %v", err)
	//}
	//
	//for _, entry := range entries {
	//	fmt.Printf("%+v\n",
	//		entry)
	//}
	//test()
	//command(func(s string) {
	//	fmt.Printf("%s %v\n", parseTime(s), s)
	//}, "logread", "-f", "grep", "|", "hostapd.*")

	httpserver.New().
		CORSMethodMiddleware().
		AddRoute(internal.NewRoute(internal.NewApi())).
		AddRoute(assets.NewRoute()).
		Done(8080)
}
