package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct{
	IP string
	Port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock sync.RWMutex

	// 消息广播的channel
	Message chan string
}



func NewServer(ip string, port int) *Server {
	server := &Server {
		IP: ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}

	return server
}

func (this *Server) Handler(conn net.Conn) {
	user := NewUser(conn, this)

	// 用户上线，将用户加入到onlineMap
	user.Online()

	// 用户获取状态管道
	isLive := make(chan bool)

	// 读取客户端消息
	go func () {
		buf := make([]byte, 4096)

		for {
			n, err := conn.Read(buf)

			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Printf("Conn Read Err:", err)
				return
			}
			
			// 用户活跃状态
			isLive <- true

			// 提取用户消息（去除末尾\n）
			msg := string(buf[:n-1])

			// 将消息广播
			user.DoMessage(msg)
		}
	}()

	// 阻塞当前handler，避免退出
	for {
		select {
		case <- isLive:
		
		case <-time.After(time.Second * 600):
			// 发送消息
			user.SendMsg("你被踢了")

			// 销毁用户资源
			close(user.C)

			// 关闭连接
			conn.Close()

			// 退出
			return
		}
	}
}

func (this *Server) Broadcast(user *User, msg string) {
	snedMsg := fmt.Sprintf("[%s]: %s",user.Name,msg)

	this.Message <- snedMsg
}

func (this *Server) ListenBroadcastMessage() {
	for {
		msg := <- this.Message

		this.mapLock.Lock()
		for _, item := range this.OnlineMap {
			item.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.IP, this.Port))
	if err != nil {
		fmt.Println("net listen err:", err)
		return
	}

	// close listen socket
	defer listener.Close()

	// 启动监听Message的goroutine
	go this.ListenBroadcastMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept err:", err)
			continue
		}

		// do handler
		go this.Handler(conn)
	}
}