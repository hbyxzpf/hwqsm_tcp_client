package hwqsm_tcp_client

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/axgle/mahonia"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Cmd string

type Channel string

const (
	Test Channel = "HWQSMWRG"
	// Tb 【神单】淘宝
	Tb Channel = "HWQSMTB"
	// TbJTWADS 【神单】淘宝社群精推版
	TbJTWADS Channel = "HWQSMJTWADS"
	// TbTMCS 【神单】猫超生活
	TbTMCS Channel = "HWQSMTMCS"
	// TbZZ 【神单】壮壮青年必买
	TbZZ Channel = "HWQSMTBZZ"
	// TbMGYP 【淘京】猫狗用品线报
	TbMGYP Channel = "HWQSMMGYP"
	// TbJMDXHJ01 【神单】淘宝全网高佣定向选品群
	TbJMDXHJ01 Channel = "HWQSMJMDXHJ01"
	// Jd 【神单】京东肉单线报版
	Jd Channel = "HWQSMJD"
	// JdJJB 【神单】京东精简线报版
	JdJJB Channel = "HWQSMJDJJB"
	// ALL 所有订阅
	ALL Channel = "HWQSMALL"
)

const (
	CmdRegisterCode      Cmd = "1"
	CmdRegisterBroadcast Cmd = "217"
	CmdMessage           Cmd = "203"
)

type Callback func(cmd CmdData)

type CmdData struct {
	Cmd         Cmd       `json:"cmd"`
	Time        string    `json:"time"`
	Code        string    `json:"code"`
	Frame       string    `json:"frame,omitempty"`
	Version     string    `json:"version,omitempty"`
	ChannelName string    `json:"channelame,omitempty"`
	Content     string    `json:"content,omitempty"`
	Broadcast   []Channel `json:"broadcast,omitempty"`
	Channel     Channel   `json:"channel,omitempty"`
	Recmd       string    `json:"recmd,omitempty"`
}

type TcpClient struct {
	Conf TcpClientConfig
	Conn *net.TCPConn
	Once sync.Once
}

type EmojiTrans struct {
	Emoji string `json:"emoji"`
}

type TcpClientConfig struct {
	Code       string
	Url        string
	Version    string
	Broadcasts []Channel
}

func NewTcpClient(conf TcpClientConfig) *TcpClient {
	return &TcpClient{Conf: conf}
}

// Start 启动方法
func (tc *TcpClient) Start(callback Callback) {
	tc.Conn = tc.connect()
	if tc.Conn == nil {
		log.Fatalln("connect failed!")
		return
	}
	log.Println("connect success!")
	for {
		buf := make([]byte, 4096)
		reqLen, err := tc.Conn.Read(buf)
		if err != nil {
			fmt.Println("Error to read message because of ", err)
			return
		}
		originContent := string(buf[:reqLen])
		if originContent == "heartbeat" {
			continue
		}
		tcpContent := tc.convertToString(originContent, "gbk", "utf8")
		log.Println(tcpContent)
		var cmd CmdData
		_ = json.Unmarshal([]byte(tcpContent), &cmd)
		switch cmd.Cmd {
		case CmdRegisterCode:
			_, _ = tc.registerCode()
			time.Sleep(1 * time.Second)
			_, _ = tc.registerBroadcast()
			time.Sleep(1 * time.Second)
			go tc.heartbeat()
		case CmdMessage:
			decodeString, err := base64.StdEncoding.DecodeString(cmd.Content)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			cmd.Content = tc.ConvertUnicodeEmoji(tc.convertToString(string(decodeString), "gbk", "utf8"))
			callback(cmd)
		}
	}
}

func (tc *TcpClient) ConvertUnicodeEmoji(text string)string {
	reg, err := regexp.Compile("(\\\\u[a-zA-z0-9]{4}){1,2}")
	if err != nil {
		return ""
	}
	for _, match := range reg.FindAllString(text, -1) {
		sJsonBytes := []byte(fmt.Sprintf(`{"emoji":"%s"}`,match))
		var emojiTrans EmojiTrans
		_ = json.Unmarshal(sJsonBytes,&emojiTrans)
		text = strings.Replace(text,match,emojiTrans.Emoji,1)
	}
	return text
}


func (tc *TcpClient) convertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}
func (tc *TcpClient) connect() *net.TCPConn {
	addr, err := net.ResolveTCPAddr("tcp", tc.Conf.Url)
	if err != nil {
		log.Println("connect error: " + err.Error())
		return nil
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	_ = conn.SetKeepAlive(false)
	return conn
}

func (tc *TcpClient) heartbeat() {
	tc.Once.Do(func() {
		log.Println("heartbeat start!")
		for {
			_, err := tc.Conn.Write([]byte("heartbeat"))
			if err != nil {
				log.Println(err.Error())
			}
			time.Sleep(20 * time.Second)
		}
	})
}

func (tc *TcpClient) registerCode() (int, error) {
	cmdData := tc.newCmdData(CmdRegisterCode)
	cmdData.Frame = "服务器"
	cmdData.Version = tc.Conf.Version
	registerBytes, _ := json.Marshal(cmdData)
	return tc.Conn.Write(registerBytes)
}

func (tc *TcpClient) registerBroadcast() (int, error) {
	cmdData := tc.newCmdData(CmdRegisterBroadcast)
	cmdData.Broadcast = tc.Conf.Broadcasts
	cmdData.Channel = ALL
	broadcastBytes, _ := json.Marshal(cmdData)
	return tc.Conn.Write(broadcastBytes)
}

func (tc *TcpClient) newCmdData(cmd Cmd) *CmdData {
	return &CmdData{
		Cmd:  cmd,
		Code: tc.Conf.Code,
		Time: strconv.FormatInt(time.Now().UnixMilli(), 10),
	}
}
