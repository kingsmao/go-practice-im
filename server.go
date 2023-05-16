package main

import (
	"fmt"
	"io"
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
func (this *Server) UserHandler(conn net.Conn) {
	//fmt.Println("链接建立成功")
	//用户上线，将用户加入onlineMap中
	user := NewUser(conn, this)
	user.Online()
	//监听用户是否活跃
	isAlive := make(chan bool)
	//接受客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			read, err := conn.Read(buf)
			if read == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}
			//提取用户的消息（去除'\n'）
			msg := string(buf[:read-1])
			//广播消息
			user.DoMessage(msg)
			//用户的任何消息都会刷新活跃状态
			isAlive <- true
		}
	}()
	//阻塞
	for {
		//两个case都会触发，如果活跃，定时器的case会触发而内容体不会执行
		select {
		case <-isAlive:
		//当前用户是活跃的，需要重置定时器
		//不需要做任何动作，为了激活select，更新下面的定时器
		case <-time.After(time.Second * 10):
			//已经超时，将当前用户强制踢掉
			user.sendMsg("超时被踢")
			//关闭资源
			close(user.C)
			//delete(this.OnlineMap,user.Name)//不需要删除，有检测在线状态方法会删除
			//关闭链接
			conn.Close()
			//退出当前handler
			return //runtime.Goexit()

		}
	}
}

func (this *Server) ReadClientMessage(conn net.Conn, user *User) {

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
		go this.UserHandler(conn) //异步处理业务
	}
}
