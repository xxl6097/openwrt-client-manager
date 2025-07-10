package main

import (
	"fmt"
	"regexp"
)

func main() {
	logLine := "Wed Jul  9 14:57:54 2025 daemon.notice hostapd: phy1-ap0: AP-STA-DISCONNECTED 7a:34:62:d5:a4:18"

	// 编译正则表达式 [3,5](@ref)
	re := regexp.MustCompile(`hostapd:\s+(phy[\w-]+?):`)

	// 提取匹配结果
	matches := re.FindStringSubmatch(logLine)
	if len(matches) < 2 {
		fmt.Println("未找到匹配字段")
		return
	}
	phyField := matches[1]         // 捕获组索引为1
	fmt.Println("解析结果:", phyField) // 输出: phy1-ap0
}
