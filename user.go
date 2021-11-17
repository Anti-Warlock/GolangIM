package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	Server *Server
}

//新建User
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		Server: server,
	}
	//每当新建一个对象,新起一个go程监听channel中的消息
	go user.ListenMsg()
	return user
}

//监听User对象中的channel,如果有消息就给对端客户端发消息
func (this *User) ListenMsg() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}

//用户上线
func (this *User) Online() {
	//将上线的用户加入到OnlineMap
	this.Server.mapLock.Lock()
	this.Server.OnlineMap[this.Name] = this
	this.Server.mapLock.Unlock()

	//当用户连接了,开始广播消息
	this.Server.BroadCast(this, "上线了")
}

//用户下线
func (this *User) Offline() {
	this.Server.mapLock.Lock()
	delete(this.Server.OnlineMap, this.Name)
	this.Server.mapLock.Unlock()

	//当用户连接了,开始广播消息
	this.Server.BroadCast(this, "下线")
}

//TODO 此处有个Bug,在终端时只能发送一个命令,其他命令就不生效了

//用户处理消息
func (this *User) DealMsg(msg string) {
	if msg == "who" {
		//通过who指令查询当前在线的用户
		this.Server.mapLock.Lock()
		for _, user := range this.Server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + "在线!\n"
			//给当前用户发消息
			this.SendMsg(onlineMsg)
		}
		this.Server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//通过rename|指令修改用户名
		newName := strings.Split(msg, "|")[1]
		//判断newName是否存在
		_, ok := this.Server.OnlineMap[newName]
		if ok {
			this.SendMsg("当前用户名已被占用!")
		} else {
			this.Server.mapLock.Lock()
			delete(this.Server.OnlineMap, this.Name)
			this.Server.OnlineMap[newName] = this
			this.Server.mapLock.Unlock()

			//修改属性名
			this.Name = newName
			this.SendMsg("您已更新用户名:" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		//给指定用户发消息

		//1.找到用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("消息格式不正确,请使用\"to|张三|你好啊"+"格式!\n")
			return
		}
		//2.获取User对象
		remoteUser, ok := this.Server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("用户不存在!\n")
			return
		}
		//3.获取消息
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("发送消息内容不能为空!\n")
			return
		}
		//4.调用目标用户的给客户端发送消息的方法
		remoteUser.SendMsg(this.Name + "对您说:" + content + "\n")
	} else {
		//广播消息
		this.Server.BroadCast(this, msg)
	}
}

//给当前用户发消息
func (this *User) SendMsg(msg string) {

	this.conn.Write([]byte(msg))
}
