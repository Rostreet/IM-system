package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	Conn   net.Conn
	Server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string, 16),
		Conn:   conn,
		Server: server,
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

func (this *User) Online() {
	//将用户加入到在线用户列表
	this.Server.mapLock.Lock()
	this.Server.OnlineMap[this.Name] = this
	this.Server.mapLock.Unlock()
	//增加消息
	this.Server.AddMessage(this, "已上线")
}

// 下线通知
func (this *User) Offline() {
	this.Server.AddMessage(this, "已下线")
	//用户下线，将用户从在线列表中删除
	this.Server.mapLock.Lock()
	delete(this.Server.OnlineMap, this.Name)
	this.Server.mapLock.Unlock()
}

// 发送消息
func (this *User) SendMsg(msg string) {
	if msg == "who" {
		//查询当前在线用户都有谁（先快照，避免持锁期间阻塞发送）
		this.Server.mapLock.RLock()
		users := make([]*User, 0, len(this.Server.OnlineMap))
		for _, user := range this.Server.OnlineMap {
			users = append(users, user)
		}
		this.Server.mapLock.RUnlock()

		for _, user := range users {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":在线..."
			this.C <- onlineMsg
		}
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := strings.Split(msg, "|")[1]
		//查询用户名是否存在
		_, ok := this.Server.OnlineMap[newName]
		if ok {
			this.C <- "当前用户名已被使用"
		} else {
			this.Server.mapLock.Lock()
			delete(this.Server.OnlineMap, this.Name)
			this.Server.OnlineMap[newName] = this
			this.Server.mapLock.Unlock()

			this.Name = newName
			this.C <- "您已更新用户名:" + this.Name
		}
	} else {
		this.Server.AddMessage(this, msg)
	}
}
