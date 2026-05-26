package room

import (
	"slices"

	"github.com/dev-au/CodeStream/internal/domain/room"
)

func mergePatch(currentText []byte, patch *room.Patch) []byte {
	switch patch.Action {
	case room.ActionAddType:
		currentText = slices.Insert(currentText, patch.StartIndex, patch.Content...)
	case room.ActionDeleteType:
		currentText = slices.Delete(currentText, patch.StartIndex, *patch.EndIndex)
	}

	return currentText
}
