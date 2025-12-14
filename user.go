package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	Conn net.Conn
}

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		Conn: conn,
	}
	//启动监听
	go user.ListenMessage()

	return user
}

// 监听当前 User 的 channel 方法，一旦监听到server 给它发送消息，就返回给 client
func (this *User) ListenMessage() {
	for {
		msg := <-this.C

		_, err := this.Conn.Write([]byte(msg + "\n"))
		if err != nil {
			// 连接写入失败，退出循环
			return
		}
	}
}
