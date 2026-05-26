package room

import "github.com/dev-au/CodeStream/internal/domain/account"

type Room struct {
	ID      string
	Owner   *account.User
	Guests  []*account.User
	Content []byte
	Patches []*Patch
}

type ActionType int8

const (
	ActionAddType ActionType = iota
	ActionDeleteType
	ActionSelectType
)

type Patch struct {
	Action     ActionType
	Content    []byte
	StartIndex int
	EndIndex   *int
}
