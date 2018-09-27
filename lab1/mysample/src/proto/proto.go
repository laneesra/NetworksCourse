package proto

import "encoding/json"

// Request -- запрос клиента к серверу.
type Request struct {
	// Поле Command может принимать три значения:
	// * "quit" - прощание с сервером (после этого сервер рвёт соединение);
	// * "add" - передача нового элемента на сервер;
	// * "peak" - просьба посчитать количество пиков в последовательности
	Command string `json:"command"`

	// Если Command == "add", в поле Data должен лежать элемент
	// в виде структуры Element.
	// В противном случае, поле Data пустое.
	Data *json.RawMessage `json:"data"`
}

// Response -- ответ сервера клиенту.
type Response struct {
	// Поле Status может принимать три значения:
	// * "ok" - успешное выполнение команды "quit" или "add";
	// * "failed" - в процессе выполнения команды произошла ошибка;
	// * "result" - количество пиков в последовательности вычислено.
	Status string `json:"status"`

	// Если Status == "failed", то в поле Data находится сообщение об ошибке.
	// Если Status == "result", в поле Data должно лежать целое число (int64)
	// В противном случае, поле Data пустое.
	Data *json.RawMessage `json:"data"`
}

type Element struct {
	// Целое число(в десятичной системе, разрешён знак).
	Value string `json:"val"`
	
	// Новая позиция элемента(целое число, не больше длины последовательности)
	Index string `json:"ind"`
}
