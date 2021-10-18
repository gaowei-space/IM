package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C: make(chan string),
		conn: conn,

		server: server,
	}

	// 启动监听user的channel消息的 goroutine
	go user.ListenMessage()

	return user
}

func (this *User) Online()  {
	// 用户上线，将用户加入到onlineMap
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	this.DoMessage("用户上线")
}

func (this *User) Offline()  {
	// 用户下线，将用户从onlineMap移除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap,this.Name)
	this.server.mapLock.Unlock()

	this.DoMessage("用户下线")
}

func (this *User) SendMsg(msg string)  {
	// 用户下线，将用户从onlineMap移除
	this.conn.Write([]byte(msg))
}

func (this *User) DoMessage(msg string)  {

	if msg == "who" {
		this.server.mapLock.Lock()
		for _, item := range this.server.OnlineMap {
			onlineMsg := fmt.Sprintf("[%s]在线\n", item.Name)
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 4 && msg[:3] == "to|" {
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg("用户名不能为空\n")
			return
		}

		toUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsg("用户不存在\n")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("消息不能为空\n")
			return
		}

		toUser.SendMsg(content + "---来自["+ this.Name + "]\n")

	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg,"|")[1]

		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsg("当前用户名已存在\n")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap,this.Name)
			this.server.OnlineMap[newName] = this
			this.server.mapLock.Unlock()

			this.Name = newName

			this.SendMsg("修改成功\n")
		}
	} else {
		this.server.Broadcast(this, msg)
	}
}

func (this *User) ListenMessage() {
	for {
		msg := <- this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}