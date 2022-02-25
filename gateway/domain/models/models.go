package models

type Data struct {
	Symbol    string
	Price     float64
	Timestamp string
}

type RequestAction string

const (
	Subscribe   = RequestAction("subscribe")
	Unsubscribe = RequestAction("unsubscribe")
)

type GatewayRequest struct {
	Action  RequestAction `json:"action"`
	Symbols []string      `json:"symbols,omitempty"`
}

type ResponseStatus string

const (
	SuccessStatus = ResponseStatus("success")
	ErrorStatus   = ResponseStatus("error")
)

type GatewayResponse struct {
	Request GatewayRequest `json:"request"`
	Status  ResponseStatus `json:"status"`
	Error   *string        `json:"error,omitempty"`
}

func (g *GatewayResponse) SetError(errorMsg string) {
	g.Error = &errorMsg
}
