package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn
}

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C: make(chan string),
		conn: conn,
	}

	// 启动监听user的channel消息的 goroutine
	go user.ListenMessage()

	return user
}

func (this *User) ListenMessage() {
	for {
		msg := <- this.C

		this.conn.Write([]byte(msg + "\n"))
	}
}