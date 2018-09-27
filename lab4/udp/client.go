		case "quit":
			for {
				t:=time.Now()
				t=t.Add(3*time.Second)
				conn.SetDeadline(t)
				if ident == strconv.Itoa(int(i)) {
					return
				}
				if send_request(conn, "quit", nil, i){
					t=time.Now()
					t=t.Add(300*time.Millisecond)
					conn.SetDeadline(t)
					get_response(conn, buf)
			 	}
		 }
package main

import (
	"flag"
	"fmt"
	"github.com/mgutz/logxi/v1"
	"github.com/skorobogatov/input"
	"net"
	"os"
	"encoding/json"
	"strconv"
	"time"
)

import "proto"

// send_request - вспомогательная функция для передачи запроса с указанной командой
// и данными. Данные могут быть пустыми (data == nil).
func send_request(conn *net.UDPConn, command string, data interface{}, ident uint) bool {
	var rawData json.RawMessage
	rawData, _ = json.Marshal(data)
	rawReq, _ := json.Marshal(&proto.Request{command, &rawData, strconv.Itoa(int(ident))})
	if _, err := conn.Write(rawReq); err != nil {
		log.Error("sending request to server", "error", err, "command", command)
		return false
	}
	return true
}
 var ident string

 // get_response - вспомогательная функция для получения ответа от сервера.
 // Данные могут быть пустыми (data == nil).
 // При правильно обработанном запросе ident содержит идентификатор запроса, иначе - сообщение об ошибке
func get_response(conn *net.UDPConn, buf []byte) {
	if bytesRead, err := conn.Read(buf); err != nil {
		log.Info("receiving answer from server")
	} else {
		var resp proto.Response
		if err := json.Unmarshal(buf[:bytesRead], &resp); err != nil {
			log.Error("cannot parse answer", "answer", buf, "error", err)
		} else {
			switch resp.Status {
			case "ok":
				ident = resp.Ident
				fmt.Printf("ok\n")
				return
			case "added":
				var elem proto.Element
				if err := json.Unmarshal(*resp.Data, &elem); err != nil {
					log.Error("cannot parse answer", "answer", resp.Data, "error", err)
					ident = "failed"
					fmt.Printf("error: malformed data field in response\n")
				} else {
					log.Info("successful interaction with server", "added", elem.Value, "in", elem.Index)
					ident = resp.Ident
					fmt.Printf("ok\n")
					return
				}
			case "failed":
				var reason string
				if err := json.Unmarshal(*resp.Data, &reason); err != nil {
					log.Error("cannot parse answer", "answer", resp.Data, "error", err)
					ident = "failed"
					fmt.Printf("error: malformed data field in response\n")
				} else {
					fmt.Printf("failed: %s\n", reason)
					ident = reason
				}
				return
			case "peak":
				var count string
				if err := json.Unmarshal(*resp.Data, &count); err != nil {
					log.Error("cannot parse answer", "answer", resp.Data, "error", err)
					fmt.Printf("error: malformed data field in response\n")
				} else {
					log.Info("successful interaction with server", "peak", count)
					fmt.Printf("result: " + count + "\n")
					ident = resp.Ident
					return
				}
			default:
				fmt.Printf("error: server reports unknown status %q\n", resp.Status)
			}
		}
	}
}

// interact - функция, содержащая цикл взаимодействия с сервером.
func interact(conn *net.UDPConn, n uint) {
	defer conn.Close()
	buf := make([]byte, 100)
	for i := uint(0); i < n; i++ {
		ident = "clear"
		fmt.Printf("command = ")
		command := input.Gets()
		switch command {
		case "add":
			var elem proto.Element
			fmt.Printf("value = ")
			elem.Value = input.Gets()
			fmt.Printf("index = ")
			elem.Index = input.Gets()
			for {
				t:=time.Now()
				t=t.Add(3*time.Second)
				conn.SetDeadline(t)
					if ident == "index out of range" || ident == "malformed data field"{
						log.Error("incorrect data", "reason", ident)
						i--
						t=t.Add(1*time.Hour)
						conn.SetDeadline(t)
						break
					}
					if ident == strconv.Itoa(int(i)) {
						t=t.Add(1*time.Hour)
						conn.SetDeadline(t)
						break
					}
					if ident != "clear" {
						t=t.Add(1*time.Hour)
						conn.SetDeadline(t)
						log.Info("sending the request again")
					}
					if send_request(conn, "add", &elem, i){
						t=time.Now()
						t=t.Add(300*time.Millisecond)
						conn.SetDeadline(t)
						get_response(conn, buf)
					}
				}
		case "peak":
		  for {
				t:=time.Now()
				t=t.Add(3*time.Second)
				conn.SetDeadline(t)
				if ident == strconv.Itoa(int(i)) {
					t=t.Add(1*time.Hour)
					conn.SetDeadline(t)
					break
				}
				if ident != "clear" {
					t=t.Add(1*time.Hour)
					conn.SetDeadline(t)
					log.Info("sending the request again")
				}
				if send_request(conn, "peak", nil, i){
					t:=time.Now()
					t=t.Add(300*time.Millisecond)
					conn.SetDeadline(t)
					get_response(conn, buf)
				}
			}
		default:
			fmt.Printf("error: unknown command\n")
			i--
			continue
		}
	}
}

func main() {
	var (
		serverAddrStr string
		n             uint
		helpFlag      bool
	)
	flag.StringVar(&serverAddrStr, "server", "127.0.0.1:6000", "set server IP address and port")
	flag.UintVar(&n, "n", 10, "set the number of requests")
	flag.BoolVar(&helpFlag, "help", false, "print options list")

	if flag.Parse(); helpFlag {
		fmt.Fprint(os.Stderr, "client [options]\n\nAvailable options:\n")
		flag.PrintDefaults()
	} else if serverAddr, err := net.ResolveUDPAddr("udp", serverAddrStr); err != nil {
		log.Error("resolving server address", "error", err)
	} else if conn, err := net.DialUDP("udp", nil, serverAddr); err != nil {
		log.Error("creating connection to server", "error", err)
	} else {
		interact(conn, n)
	}
}
package main

import (
	"flag"
	"fmt"
	"github.com/mgutz/logxi/v1"
	"github.com/skorobogatov/input"
	"net"
	"os"
	"encoding/json"
	"strconv"
	"time"
)

import "proto"

// send_request - вспомогательная функция для передачи запроса с указанной командой
// и данными. Данные могут быть пустыми (data == nil).
func send_request(conn *net.UDPConn, command string, data interface{}, id uint)  {
	var rawData json.RawMessage
	rawData, _ = json.Marshal(data)
	ident := strconv.Itoa(int(id))
	rawReq, _ := json.Marshal(&proto.Request{command, &rawData, ident})
	buf := make([]byte, 2000)
	for {
		t:=time.Now()
		t=t.Add(5*time.Second)
		conn.SetDeadline(t)
		if _, err := conn.Write(rawReq); err != nil {
			log.Error("sending request to server", "error", err)
			log.Info("sending the request again")
			continue
		}
		// Ставим время попытки приёма сообщения на 2 секунды.
		// Если за это время ответ не придёт, то заново посылаем запрос.
		t=time.Now()
		t=t.Add(300*time.Millisecond)
		conn.SetDeadline(t)
		if bytesRead, err := conn.Read(buf); err != nil {
			log.Error("receiving answer from server", "error", "timeout")
			continue
		} else {
			t=t.Add(2*time.Hour)
			conn.SetDeadline(t)
			var resp proto.Response
			if err := json.Unmarshal(buf[:bytesRead], &resp); err != nil {
				log.Error("cannot parse answer", "answer", buf, "error", err)
			} else {
				switch resp.Status {
				case "ok":
					if resp.Ident == ident {
						fmt.Printf("ok\n")
						return
					}
				case "added":
					var elem proto.Element
					if err := json.Unmarshal(*resp.Data, &elem); err != nil {
						log.Error("cannot parse answer", "answer", resp.Data, "error", err)
					} else {
						if resp.Ident == ident {
							log.Info("successful interaction with server", "added", elem.Value, "in", elem.Index)
							return
						}
					}
				case "failed":
					var reason string
					if err := json.Unmarshal(*resp.Data, &reason); err != nil {
						log.Error("cannot parse answer", "answer", resp.Data, "error", err)
					} else {
						if resp.Ident == ident {
							log.Error("failed", "reason", reason)
							return
						}
					}
				case "peak":
					var count string
					if err := json.Unmarshal(*resp.Data, &count); err != nil {
						log.Error("cannot parse answer", "answer", resp.Data, "error", err)
						fmt.Printf("error: malformed data field in response\n")
					} else {
						if resp.Ident == ident {
							log.Info("successful interaction with server", "peak", count)
							fmt.Printf("result: " + count + "\n")
							return
						}
					}
				default:
					fmt.Printf("error: server reports unknown status %q\n", resp.Status)
				}
			}
		}
	}
}

// interact - функция, содержащая цикл взаимодействия с сервером.
func interact(conn *net.UDPConn, n uint) {
	defer conn.Close()
	for i := uint(0); i < n; i++ {
		fmt.Printf("command = ")
		command := input.Gets()
		switch command {
		case "add":
			var elem proto.Element
			fmt.Printf("value = ")
			elem.Value = input.Gets()
			fmt.Printf("index = ")
			elem.Index = input.Gets()
			send_request(conn, "add", &elem, i)
		case "peak":
		  	send_request(conn, "peak", nil, i)
		default:
			fmt.Printf("error: unknown command\n")
			i--
			continue
		}
	}
}

func main() {
	var (
		serverAddrStr string
		n             uint
		helpFlag      bool
	)
	flag.StringVar(&serverAddrStr, "server", "127.0.0.1:6000", "set server IP address and port")
	flag.UintVar(&n, "n", 10, "set the number of requests")
	flag.BoolVar(&helpFlag, "help", false, "print options list")

	if flag.Parse(); helpFlag {
		fmt.Fprint(os.Stderr, "client [options]\n\nAvailable options:\n")
		flag.PrintDefaults()
	} else if serverAddr, err := net.ResolveUDPAddr("udp", serverAddrStr); err != nil {
		log.Error("resolving server address", "error", err)
	} else if conn, err := net.DialUDP("udp", nil, serverAddr); err != nil {
		log.Error("creating connection to server", "error", err)
	} else {
		interact(conn, n)
	}
}

