package main

import (
	"log"
)
import hwqsm "github.com/hbyxzpf/hwqsm_tcp_client"

func main() {
	client := hwqsm.NewTcpClient(hwqsm.TcpClientConfig{
		Code:       "80F2E8937E8FBD786C9C7316EBDE4D79185100717231652934252",
		Version:    "1.5",
		Url:        "channel.hwqsm.com:3095",
		Broadcasts: []hwqsm.Channel{hwqsm.Tb, hwqsm.Jd, hwqsm.Test},
	})
	client.Start(func(cmd hwqsm.CmdData) {
		log.Println(cmd.Content)
		//正则匹配图片
		//var re = regexp.MustCompile(`(?m)\[pic,hash=([A-Z0-9]+),wide=(\d+),high=(\d+),cartoon=[a-z]+\]\[photo=(.*)\]`)
		//for i, match := range re.FindAllStringSubmatch(cmd.Content, -1) {
		//	fmt.Println(match, "found at index", i)
		//}
	})
}
