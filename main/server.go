package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表
	OnlineMap map[string]*User
	RWLock    sync.RWMutex

	// 消息广播的channel
	ChanMessage chan string

	// 服务器操作系统
	RunTimeOs string
}

// 创建一个server对象，返回对象地址。
func NewServer(ip string, port int, runTimeOs string) (server *Server) {
	server = &Server{
		Ip:          ip,
		Port:        port,
		OnlineMap:   make(map[string]*User),
		ChanMessage: make(chan string),
		RunTimeOs:   runTimeOs,
	}
	return
}

// 监听ChanMessage，一旦有消息发送给全部的在线User
func (s *Server) ListenMessage() {
	for {
		msg := <-s.ChanMessage
		s.RWLock.Lock()
		for _, cli := range s.OnlineMap {
			// fmt.Println(key, cli)
			cli.C <- msg
		}
		s.RWLock.Unlock()
	}
}

// 广播消息到ChanMessage
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := fmt.Sprintf("[%s] %s: %s", user.Addr, user.Name, msg)
	// fmt.Println(sendMsg)
	s.ChanMessage <- sendMsg
}

// 用户上线handler
func (s *Server) Handler(conn net.Conn) {
	// 当前连接的业务
	user := NewUser(conn, s)
	// 用户上线
	user.Online()
	// 监听用户是否活跃
	isLive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buff := make([]byte, 4096)
		// 创建 GBK 到 UTF - 8 的解码器
		decoder := simplifiedchinese.GBK.NewDecoder()

		for {
			n, err := conn.Read(buff)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				user.Offline()
				return
			}
			// 提取用户的消息（去除'\n'）
			var sub int
			if s.RunTimeOs == "windows" {
				sub = 2
			} else {
				sub = 1
			}

			utf8Data, err := io.ReadAll(transform.NewReader(io.NopCloser(bytes.NewReader(buff[:n])), decoder))
			if err != nil {
				fmt.Println("Error decoding:", err)
				return
			}
			// 将转换后的 UTF - 8 字节切片转换为字符串
			str := string(utf8Data)
			fmt.Printf("Raw data: %v\n", buff[:n])
			fmt.Println("Received string:", str)

			// decoder := simplifiedchinese.GBK.NewDecoder()
			// decoder := simplifiedchinese.GB18030.NewDecoder()
			// utf8Bytes, err := io.ReadAll(transform.NewReader(bytes.NewReader(buff), decoder))
			// utf8Bytes :=
			fmt.Printf("==用户[%s]发送消息: \"%s\"\n", user.Name, string(buff[:n]))
			msg := string(buff[:n-sub])
			fmt.Printf("Msg Raw data: %v\n", buff[:n-sub])
			fmt.Printf("用户[%s]发送消息: \"%s\"\n", user.Name, msg)

			// 将得到的消息进行广播
			user.DoMessage(msg)

			// 用户发送的任意消息都表示用户是活跃的
			isLive <- true
		}
	}()

	// 当前handler阻塞
	for {
		select {
		case <-isLive:
			// 当前用户是活跃的，应该重置定时器
			// 不做任何事情，为了激活select，更新下面的定时器
		case <-time.After(time.Second * 60):
			// 执行到这里时已经超时
			// 将当前User强制下线
			user.SendMsg("超过10s未发言，被踢出。")
			// 销毁所用资源
			close(user.C)
			// 关闭连接
			conn.Close()
			// 退出当前handler
			return // 或者runtime.Goexit()
		}
	}
}

// 启动服务
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		panic(err)
	}
	// close listen socket
	defer listener.Close()
	// 启动监听ChanMessage的goroutine
	go s.ListenMessage()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		// do handler
		go s.Handler(conn)
	}
}
