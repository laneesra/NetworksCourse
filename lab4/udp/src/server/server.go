package main

import (
	"flag"
	"fmt"
	"github.com/mgutz/logxi/v1"
	"net"
	"os"
	"strconv"
	"encoding/json"
)

import "proto"

// Client - состояние клиента.
type Client struct {
	resp 	 map[int]proto.Response // Ответы на уже обработанные запросы
  seq	   []int		 // Последовательность целых чисел
	count  int	     // Количество пиков в последовательности
}

func NewClient() *Client {
	return &Client{
		resp: 	make(map[int]proto.Response),
		seq:    make([]int, 0, 100),
		count:  0,
	}
}

// serve - метод, в котором реализован цикл взаимодействия с клиентами
func serveToClients(conn *net.UDPConn) {
	clientMap := make(map[string]*Client)
	buf := make([]byte, 1000)

	for {
		if bytesRead, clientAddr, err := conn.ReadFromUDP(buf); err != nil {
			log.Error("receiving message from client", "error", err)
		} else {
				clientAddrStr := clientAddr.String()
				_, found := clientMap[clientAddrStr]
				if !found {
					log.Info("client is on",  "client", clientAddrStr)
					clientMap[clientAddrStr] = NewClient()
				}
				var req proto.Request
				if err := json.Unmarshal(buf[:bytesRead], &req); err != nil {
					log.Error("cannot parse request", "request", buf[:bytesRead], "error", err)
					respond("failed", err, "-1", clientAddr, conn)
				} else {
						id, _ := strconv.Atoi(req.Ident)
						resp, found := clientMap[clientAddrStr].resp[id]
						if found {
							log.Info("Sending the response again")
							respond(resp.Status, resp.Data, resp.Ident, clientAddr, conn)
						} else {
								switch req.Command {
								case "add":
									errorMsg := ""
									var elem proto.Element
									if req.Data == nil {
										errorMsg = "data field is absent"
									} else {
										if err := json.Unmarshal(*req.Data, &elem); err != nil {
											errorMsg = "cannot parse request"
										} else {
											if i, err := strconv.Atoi(elem.Index); err != nil {
												errorMsg = "malformed data field"
											} else {
												if i > len(clientMap[clientAddrStr].seq) || i < 0 {
													errorMsg = "index out of range"
												} else {
													if v, err := strconv.Atoi(elem.Value); err != nil {
														errorMsg = "malformed data field"
													} else {
														log.Info("performing addition", "value", elem.Value)

														insert := func(index, value int) []int {
															seq := clientMap[clientAddrStr].seq
															seq = seq[0 : len(seq)+1]
															copy(seq[index+1:], seq[index:])
															seq[index] = value
															return seq
														}

														clientMap[clientAddrStr].seq = insert(i, v)
														clientMap[clientAddrStr].count = 0

														if len(clientMap[clientAddrStr].seq) > 1 {
															if clientMap[clientAddrStr].seq[0] > clientMap[clientAddrStr].seq[1] {
																clientMap[clientAddrStr].count++
															}
															if clientMap[clientAddrStr].seq[len(clientMap[clientAddrStr].seq)-1] > clientMap[clientAddrStr].seq[len(clientMap[clientAddrStr].seq)-2] {
																clientMap[clientAddrStr].count++
															}
														}
														if len(clientMap[clientAddrStr].seq) > 2 {
															for j := 1; j < len(clientMap[clientAddrStr].seq)-1; j++ {
																if clientMap[clientAddrStr].seq[j-1] < clientMap[clientAddrStr].seq[j]  && clientMap[clientAddrStr].seq[j+1] < clientMap[clientAddrStr].seq[j]{
																	clientMap[clientAddrStr].count++
																}
															}
														}
													}
												}
											}
										}
									}
									if errorMsg == "" {
										var rawData json.RawMessage
										rawData, _ = json.Marshal(elem)
										clientMap[clientAddrStr].resp[id] = proto.Response{"added", &rawData, req.Ident}
										if respond("added", elem, req.Ident, clientAddr, conn) {
											log.Info("successful interaction with client", "added", elem.Value, "in", elem.Index, "client", clientAddrStr)

										}
									} else {
										log.Error("addition failed", "reason", errorMsg)
										respond("failed", errorMsg, req.Ident, clientAddr, conn)
									}
								case "peak":
									c := strconv.Itoa(clientMap[clientAddrStr].count)
									var rawData json.RawMessage
									rawData, _ = json.Marshal(c)
									clientMap[clientAddrStr].resp[id] = proto.Response{"peak", &rawData, req.Ident}
									if respond("peak", c, req.Ident, clientAddr, conn) {
										log.Info("successful interaction with client", "peak", c, "client", clientAddrStr)
									}
								case "quit":
									clientMap[clientAddrStr].resp[id] = proto.Response{"ok", nil, req.Ident}
									if respond("ok", nil, req.Ident, clientAddr, conn) {
										log.Info("client is off",  "client", clientAddrStr)
									}
								}
							}
			}
		}
	}
}

// respond - вспомогательный метод для передачи ответа с указанным статусом и данными
func respond(status string, data interface{}, ident string, addr *net.UDPAddr, conn *net.UDPConn) bool {
	var rawData json.RawMessage
	rawData, _ = json.Marshal(data)
	rawResp, _ := json.Marshal(&proto.Response{status, &rawData, ident})
	if _, err := conn.WriteToUDP(rawResp, addr); err != nil {
		log.Error("sending response to client", "error", err)
		return false
	}
	return true
}


func main() {
	var (
		serverAddrStr string
		helpFlag      bool
	)
	flag.StringVar(&serverAddrStr, "addr", "127.0.0.1:6000", "set server IP address and port")
	flag.BoolVar(&helpFlag, "help", false, "print options list")

	if flag.Parse(); helpFlag {
		fmt.Fprint(os.Stderr, "server [options]\n\nAvailable options:\n")
		flag.PrintDefaults()
	} else if serverAddr, err := net.ResolveUDPAddr("udp", serverAddrStr); err != nil {
		log.Error("resolving server address", "error", err)
	} else if conn, err := net.ListenUDP("udp", serverAddr); err != nil {
		log.Error("creating listening connection", "error", err)
	} else {
		log.Info("server listens incoming messages from clients")
		serveToClients(conn)
	}
}
