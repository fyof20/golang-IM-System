package main

import (
	"net"
	"strings"
)

type User struct {
	Name  string
	Adder string
	C     chan string
	conn  net.Conn

	server *Server
}

// 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAdder := conn.RemoteAddr().String()

	user := &User{
		Name:  userAdder,
		Adder: userAdder,
		C:     make(chan string),
		conn:  conn,

		server: server,
	}

	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// 用户的上线业务
func (this *User) Online() {

	//用户上线，将用户加入到OnlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	//广播当前用户上线消息

	this.server.BroadCast(this, "已上线")
}

// 用户的下线业务
func (this *User) Offline() {

	//用户下线，将用户从onlineMap中删除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	//广播当前用户下线消息

	this.server.BroadCast(this, "下线")
}

// 给当前User对应的客户端发信息
func (this *User) SendMessage(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前在线用户都有哪些

		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Adder + "]" + user.Name + ":" + "在线...\n"
			this.SendMessage(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]

		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMessage("用户名被占用" + "\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()
			//这里修改name时已解锁，会不会不安全？
			this.Name = newName
			this.SendMessage("您已更新用户名" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//消息格式 ： to|张三|消息内容

		//1.获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMessage("消息格式不正确，请使用\"to|张三|你好啊\"格式\n")
			return
		}
		//2.根据用户名 得到对方User对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMessage("改用户名不存在\n")
			return
		}
		//3.获取消息内容，通过对方的User对象将消息内容发送过去
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMessage("无消息内容，请重发\n")
			return
		}
		remoteUser.SendMessage(this.Name + "对您说:" + content)
	} else {
		this.server.BroadCast(this, msg)
	}
}

// 监听当前User channel 的方法，一旦有消息，就直接发送给对端客户端
func (this *User) ListenMessage() {
	for msg := range this.C {
		//不用range会使channel被关闭后一直接受数据使cpu空转
		this.conn.Write([]byte(msg + "\n"))
	}
}
