package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int //user's mode
}

func NewClient(serverIp string, severPort int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: severPort,
		flag:       999,
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

// 处理server回应的消息，直接显示到标准输出即可
func (client *Client) DealResponse() {
	//一但client.conn有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>请输入合理范围内的数字<<<<<<")
		return false
	}
}

// 查询当前在线用户
func (client *Client) SelectUser() {
	var sendMsg string
	sendMsg = "who" + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn write err:", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUser()
	fmt.Println(">>>>请输入聊天对象[用户名]，exit退出")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>请输入消息内容，exit退出")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to" + "|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn.write err:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>请输入聊天内容,exit退出：")
			fmt.Scanln(&chatMsg)
		}
	}
}

func (client *Client) PublicChat() {
	var chatMsg string

	fmt.Println(">>>>输入聊天内容，exit退出")

	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		//发送内容不为空
		if chatMsg != "" {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>输入聊天内容，exit退出")
		fmt.Scanln(&chatMsg)
	}
}

func (client *Client) UpadateName() bool {
	fmt.Println(">>>>请输入用户名:")
	fmt.Scanln(&client.Name)

	snedMsg := "rename|" + client.Name + "\n"

	_, err := client.conn.Write([]byte(snedMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (this *Client) Run() {
	for this.flag != 0 {
		for this.menu() != true {
		}

		//根据不同的模式处理不同的业务
		switch this.flag {
		case 1:
			//公聊模式
			fmt.Println("公聊模式选择...")
			this.PublicChat()
			break
		case 2:
			//私聊模式
			fmt.Println("私聊模式选择...")
			this.PrivateChat()
			break
		case 3:
			//更改用户名
			fmt.Println("更改用户名选择...")
			this.UpadateName()
			break
		}
	}
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

	//单独开一个goroutine去处理server的回执消息
	go client.DealResponse()

	//启动客户端业务
	client.Run()
}
