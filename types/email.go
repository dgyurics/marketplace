package types

type Email struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}
