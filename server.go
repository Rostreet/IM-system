package main

import (
	"fmt"
	"net"
	"sync"
	"time"
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
	user := NewUser(conn, this)
	user.Online()
	//监听用户是否活跃的 channel
	isLive := make(chan bool)
	//接收用户发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				//用户下线
				user.Offline()
				return
			}
			if err != nil {
				fmt.Println("conn read err:", err)
				return
			}
			//提取用户的消息（去除'\n'）
			msg := string(buf[:n-1])
			//广播消息
			user.SendMsg(msg)
			//用户的任意消息，代表当前用户是活跃的
			isLive <- true
		}
	}()
	//阻塞
	select {
	case <-isLive:
		//什么都不用做，会先判断是否活跃，如果活跃会自动进入下一个循环，重置定时器
	case <-time.After(time.Minute * 10):
		//如果进入当前循环，代表已经超时
		user.SendMsg(user.Name + " 被踢出下线（长时间未活跃）")
		//销毁用的资源
		close(user.C)
		//关闭连接
		conn.Close()
	}
}

// 监听广播，一旦有消息，就发送给全部
func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message
		//发送给全部的在线用户（先快照，避免持锁期间阻塞发送）
		this.mapLock.RLock()
		users := make([]*User, 0, len(this.OnlineMap))
		for _, user := range this.OnlineMap {
			users = append(users, user)
		}
		this.mapLock.RUnlock()

		for _, user := range users {
			user.C <- msg
		}
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
