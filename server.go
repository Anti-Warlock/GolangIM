package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

// NewServer 创建一个Server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// ListenMsg 给所有在线用户发消息
func (this *Server) ListenMsg() {
	for {
		//从Message管道中取出消息
		msg := <-this.Message

		//遍历所有的在线用户
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			//向所有的在线用户的chan中发送消息
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	//将上线的消息写入到Message chan中
	this.Message <- sendMsg
}

func (this *Server) handler(conn net.Conn) {
	//业务逻辑
	// fmt.Println("连接建立成功...")

	//新建一个上线用户的对象
	user := NewUser(conn, this)

	//用户上线
	user.Online()

	//新建用来标记当前用户是否存活的chan
	isLive := make(chan bool)

	//处理客户端发送的消息
	go func() {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)

			//没有读取到字符
			if n == 0 {
				//用户下线
				user.Offline()
				return
			}
			//err不为空并且不是文件的末尾
			if err != nil && err != io.EOF {
				fmt.Println("conn read error:", err)
				return
			}
			//去掉\n 0,n-2
			msg := string(buf[:n-1])

			//处理消息
			user.DealMsg(msg)

			//标定当前用户是活跃的
			isLive <- true
		}

	}()

	//超时强踢下线
	for {
		select {
		case <-isLive:
			//不需要做任何处理
			//激活当前的select,case条件顺序执行,自动重置定时器
		case <-time.After(time.Second * 300):
			//已经超时
			//将当前的User强制下线
			user.SendMsg("你已被踢!")
			//从OnlineMap中将该用户下线
			// delete(this.OnlineMap, user.Name)
			//销毁用户的chan
			close(user.C)
			//关闭连接
			conn.Close()
			//退出当前的handler
			return //runtime.Goexit()
		}
	}

}

//启动服务的端口
func (this *Server) Start() {
	//socket监听
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("socket listen fail...", err)
	}
	//close socket
	defer listen.Close()

	//启动监听Message的Goroutine
	go this.ListenMsg()

	for {
		//socket accept
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("socket listen fail...", err)
			continue
		}
		//新起一个go程作为handler
		go this.handler(conn)
	}
}
