package actor

type Status int

//go:generate go run golang.org/x/tools/cmd/stringer -type Status
const (
	_ Status = iota
	Padding
	Ready
	Start
	Stop
	Exited
	Error
)
