package main

import (
	"flag"
	"fmt"
	"net"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
}

func NewClient(serverIp string, severPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: severPort,
	}
	//链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, severPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	return client
}

var ServerIp string
var SeverPort int

// .client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&ServerIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&SeverPort, "port", 8888, "设置服务器端口(默认为8888)")
}

func main() {
	flag.Parse()
	client := NewClient(ServerIp, SeverPort)
	if client == nil {
		fmt.Println(">>>>>链接服务器失败...")
		return
	}

	fmt.Println(">>>>>链接服务器成功...")

	//启动客户端业务
	select {
	default:
	}
}
