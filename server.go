package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广发的channel
	Message chan string
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

//监听Message消息，一旦有消息就广播给所有用户
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		this.mapLock.Lock()
		for _, value := range this.OnlineMap {
			value.C <- msg
		}
		this.mapLock.Unlock()
	}
}

//广播消息
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

//(this *Server) 与Server进行绑定（内部方法）
//处理业务
func (this *Server) Handler(conn net.Conn) {
	//fmt.Println("链接建立成功")
	//用户上线，将用户加入onlineMap中
	user := NewUser(conn)
	this.mapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.mapLock.Unlock()

	//向所有用户广播该user的上线消息
	this.BroadCast(user, "已上线")
	//接受客户端发送的消息
	go this.ReadClientMessage(conn, user)
	//阻塞
	select {}
}

func (this *Server) ReadClientMessage(conn net.Conn, user *User) {
	go func() {
		buf := make([]byte, 4096)
		for {
			read, err := conn.Read(buf)
			if read == 0 {
				this.BroadCast(user, "下线")
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}
			//提取用户的消息（去除'\n'）
			msg := string(buf[:read-1])
			//广播消息
			this.BroadCast(user, msg)
		}
	}()
}

func (this *Server) Start() {
	//socket listener
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	//close listener socket
	defer listener.Close()
	//启动监听message的goroutine
	go this.ListenMessager()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listerner accept err:", err)
			continue
		}
		//do handler
		go this.Handler(conn) //异步处理业务
	}

}
