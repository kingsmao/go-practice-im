package main

import (
	"flag"
	"fmt"
	"net"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
}

func NewClient(serverIp string, serverPort int) *Client {
	//创建对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
	}
	//连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Printf("net.Dial err:", err)
		return nil
	}
	client.conn = conn
	return client
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址（默认127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口号（默认8888）")
}

func main() {
	//解析命令行
	flag.Parse()
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>> 服务器连接失败...")
		return
	}
	fmt.Println(">>>>>> 服务器连接成功...")

	//业务
	select {}
}
