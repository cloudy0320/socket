package main

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"strconv"
	"time"
)

var (
	name string

	flag string

	kind string
)

func main() {
	address, err := net.ResolveTCPAddr("tcp", "192.168.3.166:1200")
	checkError(err)
	conn, err := net.DialTCP("tcp", nil, address)
	checkError(err)

	inputChan := make(chan string)
	readChan := make(chan []string)
	loginChan := make(chan string)
	registerChan := make(chan string)
	stopChan := make(chan bool)

	go input()
	go heart(conn)
	go receive(conn)
	go login(conn, loginChan)
	go register()
	go input(inputChan, loginChan, registerChan)

	recvData := make([]byte, 128)
	kind = "login"
	for {
		flag = ""
		fmt.Println("1：登录，2：注册")
		_, err = fmt.Scanln(&flag)
		checkError(err)
		if !(flag == "1" || flag == "2") {
			fmt.Println("只能输入1和2")
			continue
		}
		if flag == "1" {
			inputChan <- "login"
		} else if flag == "2" {
			inputChan <- "register"
		}
		break
	}
	for {
		recvData = make([]byte, 128)
		fmt.Print("输入你的姓名:")
		_, err = fmt.Scanln(&name)
		checkError(err)
		_, err = conn.Write([]byte(request(kind + ";" + name + ";")))
		checkError(err)
		n, err := conn.Read(recvData)
		if err != nil {
			fmt.Println("好像哪里出错了")
			os.Exit(0)
		}
		if n == 0 {
			fmt.Println("服务器和你断开了链接")
			os.Exit(0)
		}
		if flag == "1" {
			if string(recvData[0]) == "1" {
				fmt.Println("您已成功登陆 " + name)
				break
			} else if string(recvData[0]) == "0" {
				fmt.Println("没有这个名称，请先去注册")
				continue
			} else if string(recvData[0]) == "2" {
				fmt.Println("这个号已经被登录了")
				continue
			}
		} else if flag == "2" {
			if string(recvData[0]) == "1" {
				fmt.Println("您已成功注册" + name)
				break
			} else if string(recvData[0]) == "0" {
				fmt.Println("这个名称已经被注册了")
				continue
			} else {
				for {
					recvData = make([]byte, 128)
					fmt.Print("输入你的姓名:")
					_, err = fmt.Scanln(&name)
					checkError(err)
					_, err = conn.Write([]byte(request(kind + ";" + name + ";")))
					checkError(err)
					for {
						recvData = make([]byte, 128)
						fmt.Print("输入你的姓名:")
						_, err = fmt.Scanln(&name)
						checkError(err)
						_, err = conn.Write([]byte(request(kind + ";" + name + ";")))
						checkError(err)
						n, err := conn.Read(recvData)
						if err != nil {
							fmt.Println("好像哪里出错了")
							os.Exit(0)
						}
						if n == 0 {
							fmt.Println("服务器和你断开了链接")
							os.Exit(0)
						}
						if flag == "1" {
							if string(recvData[0]) == "1" {
								fmt.Println("您已成功登陆 " + name)
								break
							} else if string(recvData[0]) == "0" {
								fmt.Println("没有这个名称，请先去注册")
								continue
							} else if string(recvData[0]) == "2" {
								fmt.Println("这个号已经被登录了")
								continue
							}
						} else if flag == "2" {
							if string(recvData[0]) == "1" {
								fmt.Println("您已成功注册" + name)
								break
							} else if string(recvData[0]) == "0" {
								fmt.Println("这个名称已经被注册了")
								continue
							} else {
								fmt.Println("服务器好像有点问题")
								continue
							}
						}
					}
					n, err := conn.Read(recvData)
					if err != nil {
						fmt.Println("好像哪里出错了")
						os.Exit(0)
					}
					if n == 0 {
						fmt.Println("服务器和你断开了链接")
						os.Exit(0)
					}
					if flag == "1" {
						if string(recvData[0]) == "1" {
							fmt.Println("您已成功登陆 " + name)
							break
						} else if string(recvData[0]) == "0" {
							fmt.Println("没有这个名称，请先去注册")
							continue
						} else if string(recvData[0]) == "2" {
							fmt.Println("这个号已经被登录了")
							continue
						}
					} else if flag == "2" {
						if string(recvData[0]) == "1" {
							fmt.Println("您已成功注册" + name)
							break
						} else if string(recvData[0]) == "0" {
							fmt.Println("这个名称已经被注册了")
							continue
						} else {
							fmt.Println("服务器好像有点问题")
							continue
						}
					}
				}
				fmt.Println("服务器好像有点问题")
				continue
			}
		}
	}
	for {
		flag = ""
		fmt.Println("进入公共聊天室输入1，想要私聊输入2")
		_, err = fmt.Scanln(&flag)
		checkError(err)
		if !(flag == "1" || flag == "2") {
			fmt.Println("只能输入1和2")
			continue
		}
		break
	}

	ch := make(chan int)
	if flag == "1" {
		go many(conn, ch)
		<-ch
	} else if flag == "2" {
		go one(conn, ch)
		<-ch
	}
}

func checkError(err error) {
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		reflect.ValueOf(err)
		checkError(err)
		os.Exit(1)
	}
}

func receive(conn net.Conn) {
	for {
		recvData := make([]byte, 2048)
		n, err := conn.Read(recvData)
		if err != nil {
			checkError(err)
			break
		}
		if n == 0 {
			break
		} else {
			recvStr := string(recvData[:n])
			fmt.Println(recvStr)
		}
	}
}

func one(conn net.Conn, ch chan int) {
Again:
	for {
		kind = "one"
		_, err := conn.Write([]byte(request("many" + ";" + "说" + ";" + "/list" + ";")))
		checkError(err)
		fmt.Println("请输入你要聊天的用户名")
		toName := ""
		_, err = fmt.Scanln(&toName)
		checkError(err)
		message := ""
		for {
			fmt.Scanln(&message)
			if message == "/quit" {
				continue Again
			} else if message == "/change" {
				go many(conn, ch)
				return
			} else {
				_, err = conn.Write([]byte(request(kind + ";" + message + ";" + toName + ";")))
				checkError(err)
			}
		}
	}
}

func many(conn net.Conn, ch chan int) {
	kind = "many"
	for {
		message := ""
		_, err := fmt.Scanln(&message)
		checkError(err)
		if message == "" {
			fmt.Println("不能输入为空")
			continue
		}
		if message == "/change" {
			go one(conn, ch)
			return
		} else {
			_, err = conn.Write([]byte(request(kind + ";" + name + "说" + ";" + message + ";")))
			_, err = conn.Write([]byte(request(kind + ";" + name + "说" + ";" + message + ";")))
			checkError(err)
		}
	}
}

func request(s string) string {
	msgLen := fmt.Sprintf("%04s", strconv.Itoa(len(s)))
	return msgLen + s
}

func heart(conn net.Conn) {
	for {
		_, err := conn.Write([]byte(request("heart")))
		if err != nil {
			fmt.Println("连接好像出了点问题，正在尝试重连")
			for {
				address, err := net.ResolveTCPAddr("tcp", "192.168.3.166:1200")
				checkError(err)
				conn, err = net.DialTCP("tcp", nil, address)
				fmt.Println(err)
				if err == nil {
					if name != "" {
						_, err = conn.Write([]byte(request("reconnect" + ";" + name + ";")))
						checkError(err)
					}
					break
				}
			}
		} else {
			fmt.Println("心跳检测")
		}
		time.Sleep(10 * time.Second)
	}
}

func login(conn net.Conn, loginChan <-chan string) {
	<-loginChan
	for {
		recvData := make([]byte, 128)
		fmt.Print("输入你的姓名:")
		_, err := fmt.Scanln(&name)
		checkError(err)
		_, err = conn.Write([]byte(request(kind + ";" + name + ";")))
		checkError(err)
		n, err := conn.Read(recvData)
		if err != nil {
			fmt.Println("好像哪里出错了")
			os.Exit(0)
		}
		if n == 0 {
			fmt.Println("服务器和你断开了链接")
			os.Exit(0)
		}
		if string(recvData[0]) == "1" {
			fmt.Println("您已成功登陆 " + name)
			break
		} else if string(recvData[0]) == "0" {
			fmt.Println("没有这个名称，请先去注册")
			continue
		} else if string(recvData[0]) == "2" {
			fmt.Println("这个号已经被登录了")
			continue
		}
	}
}

func register() {
	for {
		recvData = make([]byte, 128)
		fmt.Print("输入你的姓名:")
		_, err = fmt.Scanln(&name)
		checkError(err)
		_, err = conn.Write([]byte(request(kind + ";" + name + ";")))
		checkError(err)
		n, err := conn.Read(recvData)
		if err != nil {
			fmt.Println("好像哪里出错了")
			os.Exit(0)
		}
		if n == 0 {
			fmt.Println("服务器和你断开了链接")
			os.Exit(0)
		}
		if flag == "1" {
			if string(recvData[0]) == "1" {
				fmt.Println("您已成功登陆 " + name)
				break
			} else if string(recvData[0]) == "0" {
				fmt.Println("没有这个名称，请先去注册")
				continue
			} else if string(recvData[0]) == "2" {
				fmt.Println("这个号已经被登录了")
				continue
			}
		} else if flag == "2" {
			if string(recvData[0]) == "1" {
				fmt.Println("您已成功注册" + name)
				break
			} else if string(recvData[0]) == "0" {
				fmt.Println("这个名称已经被注册了")
				continue
			} else {
				for {
					recvData = make([]byte, 128)
					fmt.Print("输入你的姓名:")
					_, err = fmt.Scanln(&name)
					checkError(err)
					_, err = conn.Write([]byte(request(kind + ";" + name + ";")))
					checkError(err)
					for {
						recvData = make([]byte, 128)
						fmt.Print("输入你的姓名:")
						_, err = fmt.Scanln(&name)
						checkError(err)
						_, err = conn.Write([]byte(request(kind + ";" + name + ";")))
						checkError(err)
						n, err := conn.Read(recvData)
						if err != nil {
							fmt.Println("好像哪里出错了")
							os.Exit(0)
						}
						if n == 0 {
							fmt.Println("服务器和你断开了链接")
							os.Exit(0)
						}
						if flag == "1" {
							if string(recvData[0]) == "1" {
								fmt.Println("您已成功登陆 " + name)
								break
							} else if string(recvData[0]) == "0" {
								fmt.Println("没有这个名称，请先去注册")
								continue
							} else if string(recvData[0]) == "2" {
								fmt.Println("这个号已经被登录了")
								continue
							}
						} else if flag == "2" {
							if string(recvData[0]) == "1" {
								fmt.Println("您已成功注册" + name)
								break
							} else if string(recvData[0]) == "0" {
								fmt.Println("这个名称已经被注册了")
								continue
							} else {
								fmt.Println("服务器好像有点问题")
								continue
							}
						}
					}
					n, err := conn.Read(recvData)
					if err != nil {
						fmt.Println("好像哪里出错了")
						os.Exit(0)
					}
					if n == 0 {
						fmt.Println("服务器和你断开了链接")
						os.Exit(0)
					}
					if flag == "1" {
						if string(recvData[0]) == "1" {
							fmt.Println("您已成功登陆 " + name)
							break
						} else if string(recvData[0]) == "0" {
							fmt.Println("没有这个名称，请先去注册")
							continue
						} else if string(recvData[0]) == "2" {
							fmt.Println("这个号已经被登录了")
							continue
						}
					} else if flag == "2" {
						if string(recvData[0]) == "1" {
							fmt.Println("您已成功注册" + name)
							break
						} else if string(recvData[0]) == "0" {
							fmt.Println("这个名称已经被注册了")
							continue
						} else {
							fmt.Println("服务器好像有点问题")
							continue
						}
					}
				}
				fmt.Println("服务器好像有点问题")
				continue
			}
		}
	}
}

func input(inputChan <-chan string, loginChan chan<- string, registerChan chan<- string) {
	select {
	case str := <-inputChan:
		if str == "login" {
			loginChan <- "1"
		} else if str == "register" {
			registerChan <- "1"
		}
	}
}
