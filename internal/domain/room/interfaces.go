package room

type Repository interface {
	CreateRoom(room *Room) error
	DeleteRoom(roomID string) error
	GetRoom(roomID string) (*Room, error)
	ExistRoom(roomID string) (bool, error)
	UpdateRoomContent(roomID string, content []byte) error
}

type Cache interface {
	AddPatch(roomID string, patch *Patch) error
	ListPatch(roomID string) ([]*Patch, error)
	ClearPatch(roomID string) error
}