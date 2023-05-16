package main

type Server struct {
	Ip   string
	Port int
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}
	return server
}

func (this *Server) Start() {
	//socket listen
	
	//accept

	//do handler

	//close listen socket
}
