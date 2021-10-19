package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp string
	ServerPort int
	Name string
	conn net.Conn
	flag int // 当前用户模式
}

func NewClient(ip string, port int) *Client {
	// 创建客户端
	client := &Client {
		ServerIp: ip,
		ServerPort: port,
		flag: 999,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}

	client.conn = conn

	return client
}

func (client *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.修改用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("请输入0-3的数字")
		return false
	}
}

func (client *Client) updateName() bool {
	fmt.Println("请输入用户名：")
	fmt.Scanln(&client.Name)

	sendMsg := "rename|" + client.Name + "\n"

	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("write msg err:", err)
		return false
	}

	return true
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {

		}

		switch client.flag {
		case 1:
			client.PublicChat()
			break
		case 2:
			client.PrivateChat()
			break
		case 3:
			client.updateName()
			break
		}
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUsers()

	fmt.Println("请输入用户名（exit退出）：")
	fmt.Scanln(&remoteName)	

	if remoteName != "exit" {
		fmt.Println("请输入聊天内容（exit退出）：")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			client.SendMsg("to|"+remoteName+"|"+chatMsg+"\n\n")
	
			chatMsg = ""
			fmt.Println("请输入聊天内容（exit退出）")
			fmt.Scanln(&chatMsg)
		}
		client.SelectUsers()
	
		fmt.Println("请输入用户名（exit退出）：")
		fmt.Scanln(&remoteName)	
	}
}

func (client *Client) SelectUsers() {
	client.SendMsg("who")
}

func (client *Client) SendMsg(msg string) bool {
	if len(msg) == 0 {
		fmt.Println("消息不能为空")
		return false
	}

	sendMsg := msg + "\n"

	_, err := client.conn.Write([]byte(sendMsg))

	if err != nil {
		fmt.Println("发送消息异常", err)
		return false
	}
	return true
}

func (client *Client) PublicChat() {
	var chatMsg string

	fmt.Println("请输入聊天内容(输入exit退出)：")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"

			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("发送失败:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println("请输入聊天内容：")
		fmt.Scanln(&chatMsg)
	}
}

// 阻塞监听服务端响应，并输出至命令行标准输出
func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.conn)
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口")
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient(serverIp, serverPort)

	if client == nil {
		fmt.Println("创建client出错")
		return
	}

	fmt.Println("创建client成功")

	go client.DealResponse()

	client.Run()
}