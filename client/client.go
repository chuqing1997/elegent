package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	// 当前客户端的模式
	flag int
}

func NewClient(serverIp string, serverPort int) (client *Client) {
	client = &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       -1,
	}
	// 连接Server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn
	return
}

func (c *Client) UpdateName() bool {
	fmt.Println(">>>>>>>请输入用户名:")
	fmt.Scanln(&c.Name)
	sendMsg := fmt.Sprintf("rename|%s\r\n", c.Name)
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error:", err)
		return false
	}
	return true
}

func (c *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>>>>>>请输入群聊内容,exit退出:")
	for chatMsg != "exit" {
		chatMsg = ""
		fmt.Scanln(&chatMsg) // 会被空格影响中断
		// reader := *bufio.NewReader(os.Stdin)
		// var err1 error
		// chatMsg, err1 = reader.ReadString('\n')
		// if err1 != nil {
		// 	fmt.Println("reader.ReadString error:", err1)
		// 	break
		// }
		_, err := c.conn.Write([]byte(chatMsg + "\r\n"))
		if err != nil {
			fmt.Println("c.conn.Write error:", err)
			break
		}
	}
}

// 查询在线用户
func (c *Client) SelectUsers() {
	sendMsg := "who\r\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("SelectUsers.conn.Write error:", err)
	}
}

// 私聊模式
func (c *Client) PrivateChat() {
	var objectName, chatMsg string
	for {
		c.SelectUsers()
		fmt.Println(">>>>>>>请输入聊天对象名称,exit退出")
		fmt.Scanln(&objectName)
		if objectName == "exit" {
			break
		}
		for {
			fmt.Println(">>>>>>>请输入消息内容,exit退出")
			fmt.Scanln(&chatMsg)
			if chatMsg == "exit" {
				break
			}
			// 消息不为空则发送
			if len(chatMsg) > 0 {
				sendMsg := fmt.Sprintf("to|%s|%s\r\n", objectName, chatMsg)
				_, err := c.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("PrivateChat conn.Write error:", err)
				}
			}
		}
	}
}

func (c *Client) DealResponse() {
	// 永久阻塞等待c.conn套接字里面的数据,复制给os.Stdout
	io.Copy(os.Stdout, c.conn)
	// 等价于下面
	// for {
	// 	buff := make([]byte, 4086)
	// 	c.conn.Read(buff)
	// 	fmt.Println(string(buff))
	// }
}

func (c *Client) menu() bool {
	var flag int
	fmt.Println("1.群聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")
	_, err := fmt.Scanln(&flag)
	if err != nil {
		fmt.Println("menu Scanln error:", err)
		// panic(err)
	}

	if flag >= 0 && flag <= 3 {
		c.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>>请输入合法范围内的数字", flag, "不是合理操作")
		return false
	}

}

func (c *Client) Run() {
	for c.flag != 0 {
		c.menu()
		switch c.flag {
		case 1:
			c.PublicChat()
		case 2:
			c.PrivateChat()
		case 3:
			c.UpdateName()
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "服务器IP地址")
	flag.IntVar(&serverPort, "port", 8888, "服务器端口")
}

func main() {
	// 命令行解析
	flag.Parse()
	client := NewClient("127.0.0.1", 8888)
	if client == nil {
		fmt.Println(">>>>>>>连接服务器失败...")
		return
	}
	fmt.Println(">>>>>>>连接服务器成功...")

	// 处理socket发送来的消息
	go client.DealResponse()

	// 启动客户端的业务
	client.Run()
	defer client.conn.Close()
}
