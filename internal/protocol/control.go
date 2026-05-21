package protocol

type ControlMessage struct {
	Type string `json:"type"`
	ID   string `json:"id,omitempty"`
	Pos  int    `json:"pos,omitempty"`
	Next string `json:"next,omitempty"`
}
