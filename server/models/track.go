package models

type Track struct {
	Name     string `json:"name"`
	length   int    `json:"length"`
	position int    `json:"position"`
	data     []byte `json:"data"`
}
