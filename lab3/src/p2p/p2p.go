package main

import (
	"encoding/json"
	"fmt"
	"github.com/skorobogatov/input"
	"net"
	"github.com/mgutz/logxi/v1"
	"math"
)

var STATE int = 1
var MyAddr string = ""
var encoder *json.Encoder = nil

// Request -- сообщение для другого пира
type Request struct {
	// В поле Data лежит структура MyStr
	Data *json.RawMessage `json:"data"`
}

type MyStr struct {
	IP string //адрес первого запроса
	Sum int //общая сумма
}
// Peer - соединение с другим пиром
type Peer struct {
	logger log.Logger    // Объект для печати логов
	conn   *net.TCPConn  // Объект TCP-соединения
	enc    *json.Encoder // Объект для кодирования и отправки сообщений
	dec    *json.Decoder // Объект для декодирования сообщений
}

// NewPeer - конструктор объекта пира, принимает в качестве параметра
// объект TCP-соединения.
func NewPeer(conn *net.TCPConn) *Peer { 
	return &Peer{
		logger: log.New(fmt.Sprintf("peer %s", conn.RemoteAddr().String())),
		conn:   conn,
		enc:    json.NewEncoder(conn),
		dec:    json.NewDecoder(conn),
	}
}

// send_request - вспомогательная функция для передачи запроса с указанной командой
// и данными. Данные могут быть пустыми (data == nil).
func send_request(encoder *json.Encoder, data interface{}) {
	var raw json.RawMessage
	raw, _ = json.Marshal(data)
	encoder.Encode(&Request{&raw})
}

// interact - функция, содержащая цикл взаимодействия с пользователем.
func interact(conn *net.TCPConn) {
	encoder := json.NewEncoder(conn)
	for {
        // Чтение команды из стандартного потока ввода
		var command string
		command = input.Gets()

		switch command {
			case "guess 0":
				STATE = 0
				send_request(encoder, MyStr{MyAddr, 0})			
			case "guess 1":
				STATE = 1
				send_request(encoder, MyStr{MyAddr, 0})
			default:
				fmt.Printf("error: unknown command\n")
		}
	}
}

	
// serve - метод, в котором реализован цикл взаимодействия с пиром.
// Подразумевается, что метод serve будет вызаваться в отдельной go-программе.
func (peer *Peer) serve() {
	defer peer.conn.Close()
	for {
		var req Request
		if err := peer.dec.Decode(&req); err != nil {
			peer.logger.Error("cannot decode message", "reason", err)
			break
		} else {
			peer.logger.Info("received message")
			peer.handleRequest(&req)
		}
	}
}

// handleRequest - метод обработки запроса от пира.
func (peer *Peer) handleRequest(req *Request) {
		errorMsg := ""
		if req.Data == nil {
			errorMsg = "data field is absent"
		} else {
			g := new(MyStr)
			if err := json.Unmarshal(*req.Data, g); err != nil {
				errorMsg = "malformed data field"
			} else {
				if MyAddr == g.IP {
					if STATE == g.Sum {
						fmt.Printf("Correct\n")
					} else {
						fmt.Printf("Wrong\n")
					}
				} else {
					send_request(encoder, MyStr{g.IP,	int(math.Mod(float64(STATE + g.Sum), 2))})
				}
			}
		}
		if errorMsg == "" {
			peer.logger.Info("information from peer added succesfully")
		} else {
			peer.logger.Error("add failed", "reason", errorMsg)
		}
}

func listen(addrStr string) {
	var listener *net.TCPListener
	defer listener.Close()

    // Разбор адреса, строковое представление которого находится в переменной addr.
	if addr, err := net.ResolveTCPAddr("tcp", addrStr); err != nil {
		log.Error("address resolution failed", "address", addrStr)
	} else {
		log.Info("resolved TCP address", "address", addr.String())

        // Инициация слушания сети на заданном адресе.
		if listener, err = net.ListenTCP("tcp", addr); err != nil {
			log.Error("listening failed", "reason", err)
		} else {
            // Цикл приёма входящих соединений.
			for {
				if conn, err := listener.AcceptTCP(); err != nil {
					log.Error("cannot accept connection", "reason", err)
				} else {
					log.Info("accepted connection", "address", conn.RemoteAddr().String())

                    // Запуск go-программы для обслуживания клиентов.
					go NewPeer(conn).serve()
				}
			}
		}
	}
}

func main() {
	var ps [5]string
	ps[0] = "127.0.0.1:6000"
	ps[1] = "127.0.0.1:6001"
	ps[2] = "127.0.0.1:6002"
	ps[3] = "127.0.0.1:6003"
	ps[4] = "127.0.0.1:6004"

	var i, j int
	fmt.Scan(&i)
	fmt.Scan(&j)

	MyAddr = ps[i]
	ToAddr := ps[j]

  	go listen(MyAddr)

	// Разбор адреса, установка соединения с родителем и запуск отдельного
	// отдельного процеса для прослушивания его сообщений
	for {
		if addr, err := net.ResolveTCPAddr("tcp", ToAddr); err != nil {
			fmt.Printf("error: %v\n", err)
		} else if conn, err := net.DialTCP("tcp", nil, addr); err == nil {
			encoder = json.NewEncoder(conn)
      		interact(conn)  						
		}
	}
}

