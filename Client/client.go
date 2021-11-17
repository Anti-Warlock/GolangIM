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
	flag       int
}

func NewClient(ip string, port int) *Client {
	//创建客户端对象
	client := &Client{
		ServerIp:   ip,
		ServerPort: port,
		flag:       999,
	}

	//连接客户端
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Println("net dial error!", err)
		return nil
	}

	client.conn = conn
	return client
}

//处理Server的回执消息,显示到标准输出
func (this *Client) DealResponse() {
	//无限阻塞,从conn中读取数据打印到标准输出中
	io.Copy(os.Stdout, this.conn)

	//等价于以下代码
	// for {
	// 	buf := make([]byte,4096)
	// 	this.conn.Read(buf)
	// 	fmt.Println(buf)
	// }
}

//查询当前所有的用户
func (this *Client) getUsers() {
	sendMsg := "who\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn write err!", err)
		return
	}
}

//私聊
func (this *Client) PirvateChat() {
	var userName string
	var content string
	this.getUsers()
	fmt.Println("请输入聊天对象的ID,exit退出")
	fmt.Scanln(&userName)
	for userName != "exit" {
		fmt.Println("请输入聊天内容,exit退出")
		fmt.Scanln(&content)
		for content != "exit" {
			if len(content) != 0 {
				sendMsg := "to|" + userName + "|" + content + "\n"
				_, err := this.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn write error!", err)
					break
				}
			}
			content = ""
			fmt.Println("请输入聊天内容,exit退出")
			fmt.Scanln(&content)
		}
		this.getUsers()
		fmt.Println("请输入聊天对象的ID,exit退出")
		fmt.Scanln(&userName)
	}
}

//广播
func (this *Client) BoardCast() {
	var chatMsg string
	fmt.Println("请输入消息,输入exit退出:")
	fmt.Scanln(&chatMsg)
	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := this.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn write err!", err)
				break
			}
		}
		//继续让用户去输入消息
		chatMsg = ""
		fmt.Println("请输入消息,输入exit退出:")
		fmt.Scanln(&chatMsg)
	}
}

//修改用户名
func (this *Client) ChangeName() bool {
	fmt.Println("请输入新的用户名:")
	fmt.Scanln(&this.Name)

	sendMsg := "rename|" + this.Name + "\n"
	_, err := this.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn wirte error!", err)
		return false
	}
	return true

}

func (this *Client) Run() {
	for this.flag != 0 {
		for this.menu() != true {

		}
		switch this.flag {
		case 1:
			//广播
			this.BoardCast()
		case 2:
			//私聊
			this.PirvateChat()
		case 3:
			//改名
			this.ChangeName()
		}
	}
}

func (this *Client) menu() bool {
	var flag int
	fmt.Println("1.广播")
	fmt.Println("2.私聊")
	fmt.Println("3.改名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		this.flag = flag
		return true
	} else {
		fmt.Println("请输入合法范围内的数字")
		return false
	}
}

var ServerIp string
var ServerPort int

func init() {
	flag.StringVar(&ServerIp, "ip", "127.0.0.1", "设置服务器IP地址(默认127.0.0.1)")
	flag.IntVar(&ServerPort, "port", 8888, "设置服务器端口(默认8888)")
}

func main() {
	flag.Parse()

	client := NewClient(ServerIp, ServerPort)

	if client == nil {
		fmt.Println("连接服务器失败!")
		return
	}
	fmt.Println("连接服务器成功!")

	go client.DealResponse()

	//客户端业务
	client.Run()
}
