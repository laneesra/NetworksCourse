package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mgutz/logxi/v1"
	"net"
	"strconv"
)

import "proto"

// Client - состояние клиента.
type Client struct {
	logger log.Logger    // Объект для печати логов
	conn   *net.TCPConn  // Объект TCP-соединения
	enc    *json.Encoder // Объект для кодирования и отправки сообщений
	seq	   []int64		 // Последовательность целых чисел
	count  int64         // Количество пиков в последовательности
}

// NewClient - конструктор клиента, принимает в качестве параметра
// объект TCP-соединения.
func NewClient(conn *net.TCPConn) *Client {
	return &Client{
		logger: log.New(fmt.Sprintf("client %s", conn.RemoteAddr().String())),
		conn:   conn,
		enc:    json.NewEncoder(conn),
		seq:    make([]int64, 0, 100),
		count:  0,
	}
}

// serve - метод, в котором реализован цикл взаимодействия с клиентом.
// Подразумевается, что метод serve будет вызаваться в отдельной go-программе.
func (client *Client) serve() {
	defer client.conn.Close()
	decoder := json.NewDecoder(client.conn)
	for {
		var req proto.Request
		if err := decoder.Decode(&req); err != nil {
			client.logger.Error("cannot decode message", "reason", err)
			break
		} else {
			client.logger.Info("received command", "command", req.Command)
			if client.handleRequest(&req) {
				client.logger.Info("shutting down connection")
				break
			}
		}
	}
}

// handleRequest - метод обработки запроса от клиента. Он возвращает true,
// если клиент передал команду "quit" и хочет завершить общение.
func (client *Client) handleRequest(req *proto.Request) bool {
	switch req.Command {
	case "quit":
		client.respond("ok", nil)
		return true
	case "add":
		errorMsg := ""
		if req.Data == nil {
			errorMsg = "data field is absent"
		} else {
			var elem proto.Element
			if err := json.Unmarshal(*req.Data, &elem); err != nil {
				errorMsg = "malformed data field"
			} else {
				if i, err := strconv.ParseInt(elem.Index, 10, 64); err != nil {
					errorMsg = "malformed data field"
				} else { 
					if i > int64(len(client.seq)) || i < 0 {
						errorMsg = "index out of range"
					} else { 
						if v, err := strconv.ParseInt(elem.Value, 10, 64); err != nil {
							errorMsg = "malformed data field"
						} else {
							client.logger.Info("performing addition", "value", elem.Value)
							
							insert := func(index int, value int64) []int64 {
								seq := client.seq
								seq = seq[0 : len(seq)+1]
								copy(seq[index+1:], seq[index:])
								seq[index] = value
								return seq
							}
							
							isPeak := func(l, m, r int) bool {
								if l >= 0 &&  r < len(client.seq) && client.seq[l] > client.seq[m]  && client.seq[r] > client.seq[m] {
									return true
								} 
								return false
							}
							
							j := int(i)
							client.seq = insert(j, v)
							
							if isPeak(j-1, j, j+1) {
								client.count++
							}
							if isPeak(j-2, j-1, j) {
								client.count++
							}
							if isPeak(j-2, j-1, j+1) {
								client.count--
							}
							if isPeak(j, j+1, j+2) {
								client.count++
							}
							if isPeak(j-1, j+1, j+2) {
								client.count--
							}
						}
					}
				}
			}
		}
		if errorMsg == "" {
			client.respond("ok", nil)
		} else {
			client.logger.Error("addition failed", "reason", errorMsg)
			client.respond("failed", errorMsg)
		}
	case "peak":
		c := strconv.FormatInt(client.count, 10) 
		client.respond("result", c)
	default:
		client.logger.Error("unknown command")
		client.respond("failed", "unknown command")
	}
	return false
}

// respond - вспомогательный метод для передачи ответа с указанным статусом
// и данными. Данные могут быть пустыми (data == nil).
func (client *Client) respond(status string, data interface{}) {
	var raw json.RawMessage
	raw, _ = json.Marshal(data)
	client.enc.Encode(&proto.Response{status, &raw})
}

func main() {
	// Работа с командной строкой, в которой может указываться необязательный ключ -addr.
	var addrStr string
	flag.StringVar(&addrStr, "addr", "127.0.0.1:6000", "specify ip address and port")
	flag.Parse()

	// Разбор адреса, строковое представление которого находится в переменной addrStr.
	if addr, err := net.ResolveTCPAddr("tcp", addrStr); err != nil {
		log.Error("address resolution failed", "address", addrStr)
	} else {
		log.Info("resolved TCP address", "address", addr.String())

		// Инициация слушания сети на заданном адресе.
		if listener, err := net.ListenTCP("tcp", addr); err != nil {
			log.Error("listening failed", "reason", err)
		} else {
			// Цикл приёма входящих соединений.
			for {
				if conn, err := listener.AcceptTCP(); err != nil {
					log.Error("cannot accept connection", "reason", err)
				} else {
					log.Info("accepted connection", "address", conn.RemoteAddr().String())

					// Запуск go-программы для обслуживания клиентов.
					go NewClient(conn).serve()
				}
			}
		}
	}
}
