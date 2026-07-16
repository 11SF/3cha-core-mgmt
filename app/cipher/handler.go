package cipher

type HandlerConfig struct {
}

type handler struct {
}

func NewHandler(cfg HandlerConfig) *handler {
	return &handler{}
}

type CipherResponse struct {
	Result string `json:"result"`
}
