package bitmex

type Command struct {
	Op   Operation `json:"op"`
	Args []string  `json:"args"`
}

type Operation string

const (
	subscribe   = Operation("subscribe")
	unsubscribe = Operation("unsubscribe")
)

type SymbolInfo struct {
	Symbol    string  `json:"symbol"`
	LastPrice float64 `json:"lastPrice"`
	AskPrice  float64 `json:"askPrice"`
	MarkPrice float64 `json:"markPrice"`
	Timestamp string  `json:"timestamp"`
}

type SubscribeResponse struct {
	Table  string       `json:"table"`
	Action string       `json:"action"`
	Data   []SymbolInfo `json:"data"`
}

type CommandResponse struct {
	Subscribe string  `json:"subscribe"`
	Success   bool    `json:"success"`
	Error     *string `json:"error"`
	Request   Command `json:"request"`
}

type ConnectResponse struct {
	Info      string `json:"info"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
	Docs      string `json:"docs"`
	Limit     Limit  `json:"limit"`
}

type Limit struct {
	Remaining int `json:"remaining"`
}
