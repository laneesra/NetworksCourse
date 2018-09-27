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

// interact - функция взаимодействия с сервером (отдает запросы с указанной командой
// и данными, также получает от сервера ответы на запрос). Данные могут быть пустыми (data == nil).
// По истечению 2 секунд, если ответ от сервера не был получен, отправляет запрос снова.
func interact(conn *net.UDPConn, command string, data interface{}, id uint)  {
	var rawData json.RawMessage
	rawData, _ = json.Marshal(data)
	ident := strconv.Itoa(int(id))
	rawReq, _ := json.Marshal(&proto.Request{command, &rawData, ident})
	buf := make([]byte, 2000)
	for {
		t:=time.Now()
		t=t.Add(3*time.Second)
		conn.SetDeadline(t)
		if _, err := conn.Write(rawReq); err != nil {
			log.Error("sending request to server", "error", err)
			log.Info("sending the request again")
			continue
		}
		t=time.Now()
		t=t.Add(2000*time.Millisecond)
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
						log.Info("client is off")
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
					} else {
						if resp.Ident == ident {
							log.Info("successful interaction with server", "peak", count)
							fmt.Printf("result: " + count + "\n")
							return
						}
					}
				default:
					log.Error("server reports unknown status %q\n", resp.Status)
				}
			}
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
				interact(conn, "add", &elem, i)
			case "peak":
			  interact(conn, "peak", nil, i)
			case "quit":
				interact(conn, "quit", nil, i)
					return
			default:
				log.Error("unknown command")
				i--
				continue
			}
		}
	}
}
