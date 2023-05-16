package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

//创建用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr, //默认值
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	//启动监听当前user channel消息的goroutine
	go user.ListenMessage()
	return user
}

//监听当前user channle方法，一旦有消息就直接发送给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n")) //向客户端发送消息
	}
}

//用户上线
func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	//向所有用户广播该user的上线消息
	this.server.BroadCast(this, "已上线")
}

//用户下线
func (this *User) Offline() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	//向所有用户广播该user的下线消息
	this.server.BroadCast(this, "已下线")
}

//用户处理消息业务
func (this *User) DoMessage(msg string) {
	if msg == "who" {
		//查询当前都哪些在线用户
		for _, onlineUser := range this.server.OnlineMap {
			onlineMsg := "[" + onlineUser.Addr + "]" + onlineUser.Name + ":" + "在线...\n"
			this.sendMsg(onlineMsg)
		}
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式 "rename|李四"
		newName := strings.Split(msg, "|")[1]
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.sendMsg("当前用户名被占用！")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[newName] = this
			this.Name = newName
			this.server.mapLock.Unlock()
			this.sendMsg("用户名更新成功！当前用户名为：" + this.Name + "\n")
		}
	} else {
		this.server.BroadCast(this, msg)
	}
}

//给当前用户发送消息
func (this *User) sendMsg(msg string) {
	this.conn.Write([]byte(msg))
}
