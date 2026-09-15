package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

// 创建一个server接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听Message广播消息channel的goroutine,一旦有消息就发送给全部在线的用户
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message
		//将msg发送给全部在线的用户
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 广播消息的方法
func (this *Server) BroadCast(user *User, msg string) {
	SendMsg := "[" + user.Adder + "]" + user.Name + ":" + msg
	this.Message <- SendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//...当前链接的业务
	//fmt.Println("链接建立成功")

	user := NewUser(conn, this)

	user.Online()

	//监听用户是否活跃的channel，true代表活跃，false代表下线或断开
	isLive := make(chan bool, 1)

	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				break
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				break
			}

			//提取用户消息(去除\n)
			msg := string(buf[:n-1])

			//将得到的消息进行广播
			user.DoMessage(msg)

			//用户的任意消息，代表当前用户是一个活跃的
			isLive <- true
		}
		isLive <- false
	}()

	//当前handler阻塞
	for {
		select {
		case live := <-isLive:
			//当前用户是活跃的，应该重置定时器
			//不做任何事情，为了激活select,更新下面的定时器

			if live == false {
				//客户端已断开，和超时踢人走同一套清理
				user.Offline()
				close(user.C)
				conn.Close()
				return
			}

		case <-time.After(time.Minute * 5):
			//已经超时
			//将当前的User强制关闭
			user.SendMessage("你被踢了")

			//下线
			user.Offline()

			//销毁用的资源
			close(user.C)

			//关闭连接
			conn.Close()

			//退出当前Handler
			return //runtime.Goexit()
		}
	}
}

// 启动服务器的接口
func (this *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	//close listen socket
	defer listener.Close()

	//启动监听Message的goroutine
	go this.ListenMessage()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		//do handler
		go this.Handler(conn)
	}
}
