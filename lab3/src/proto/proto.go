package proto

import "encoding/json"

// Request -- сообщение для другого пира
type Request struct {
	// В поле Data лежит структура MyStr
	Data *json.RawMessage `json:"data"`
}

type MyStr struct {
	IP string //адрес первого запроса
	Sum int //общая сумма
}
