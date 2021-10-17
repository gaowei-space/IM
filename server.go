package main

import (
	"fmt"
	"net"
	"sync"
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
	user := NewUser(conn)

	// 用户上线，将用户加入到onlineMap
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	// 广播当前用户上线消息
	this.Broadcast(user, "已上线")

	// 阻塞当前handler，避免退出
	select {}
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