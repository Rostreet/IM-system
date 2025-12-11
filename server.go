package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

func (this *Server) Handler(conn net.Conn) {
	//...当前链接的业务
	fmt.Println("连接成功")
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
	//accept
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept failed:", err)
			continue
		}
		//handler
		go this.Handler(conn)
	}
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}
	return server
}
