package room

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"

	"github.com/dev-au/CodeStream/internal/domain/account"
	"github.com/dev-au/CodeStream/internal/domain/room"
	"github.com/dev-au/CodeStream/pkg/customerrors"
	"github.com/dev-au/CodeStream/pkg/logs"
)

type UseCase struct {
	roomRepo    room.Repository
	accountRepo account.Repository
	patchCache  room.Cache
}

func (uc *UseCase) generateRoomID() string {
	roomInt := 100000 + rand.Intn(1000000)
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprint(rune(roomInt))))
}

func (uc *UseCase) CreateRoom(ctx context.Context, owner *account.User) (err error) {
	roomEntity := room.Room{
		ID:    "",
		Owner: owner,
	}
	for {
		roomEntity.ID = uc.generateRoomID()
		err = uc.roomRepo.CreateRoom(&roomEntity)

		if errors.Is(err, customerrors.ErrDataDuplicated) {
			logs.WarnCtx(ctx, "Room duplicated roomID", roomEntity.ID)
			err = nil
			continue
		}
		if err == nil {
			break
		}
	}

	if err != nil {
		return
	}

	return
}

func (uc *UseCase) GetRoomFullDetail(ctx context.Context, roomID string) (room *room.Room, err error) {
	room, err = uc.roomRepo.GetRoom(roomID)
	if err != nil {
		return
	}
	room.Patches, err = uc.patchCache.ListPatch(roomID)
	if err != nil {
		return nil, err
	}
	return
}

func (uc *UseCase) AddPatch(ctx context.Context, roomID string, patch *room.Patch) (err error) {
	// If necessary, check room id existence
	return uc.patchCache.AddPatch(roomID, patch)
}

func (uc *UseCase) CompactPatches(ctx context.Context, roomID string) (err error) {
	room, err := uc.roomRepo.GetRoom(roomID)
	if err != nil {
		return
	}
	listPatches, err := uc.patchCache.ListPatch(roomID)
	if err != nil {
		return
	}
	currentText := room.Content
	for i := 0; i < len(listPatches); i++ {
		patch := listPatches[i]
		currentText = mergePatch(currentText, patch)
	}
	room.Content = currentText
	err = uc.roomRepo.UpdateRoomContent(roomID, room.Content)
	return err
}

func NewUseCase(roomRepo room.Repository, accoutRepo account.Repository) *UseCase {
	return &UseCase{
		roomRepo:    roomRepo,
		accountRepo: accoutRepo,
	}
}
