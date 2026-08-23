package models

type CommonMsg struct {
	Flag string `json:"flag"`
	Body any    `json:"body"`
}

type GlobalMsg struct {
	Flag string `json:"flag"`
	Body any    `json:"body"`
}
