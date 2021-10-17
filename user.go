package main

import "net"

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

func (this *User) DoMessage(msg string)  {
	this.server.Broadcast(this, msg)
}

func (this *User) ListenMessage() {
	for {
		msg := <- this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}