package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int
	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	//消息广播的 channel
	Message chan string
}

func (this *Server) AddMessage(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//...当前链接的业务
	//用户上线
	user := NewUser(conn)
	//将用户加入到在线用户列表
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()
	//增加消息
	this.AddMessage(user, "已上线")

	//阻塞
	select {}
}

// 监听广播，一旦有消息，就发送给全部
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message
		//发送给全部的在线用户
		this.mapLock.Lock()
		for _, user := range this.OnlineMap {
			user.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) Start() {
	// socket listening
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("listen failed:", err)
		return
	}
	//close listen socket
	defer listener.Close()

	go this.ListenMessage()
	//accept 之后代表有用户登录了
	for {
		//阻塞式等待新的客户端 TCP 连接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept failed:", err)
			continue
		}
		//handler
		//每有一个新的链接，就启动一个独立的goroutine进行处理
		go this.Handler(conn)
	}
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}
