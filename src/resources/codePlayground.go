package resources

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/redis/go-redis/v9"
)

var validOperations = map[string]bool{
	"add":     true,
	"remove":  true,
	"replace": true,
}

func (interview *Interview) AddCodePatch(patch CodePatch) error {
	c := interview.Cache
	if !validOperations[patch.Operation] {
		return fmt.Errorf("invalid patch operation: %s", patch.Operation)
	}

	if patch.StartPos < 0 || (patch.Operation != "add" && patch.EndPos <= patch.StartPos) {
		return fmt.Errorf("invalid position range")
	}

	return c.Client.Watch(c.Ctx, func(tx *redis.Tx) error {
		currentVersion, err := tx.Get(c.Ctx, interview.VersionCacheKey).Int64()
		if err != nil && err != redis.Nil {
			return err
		}

		// Check if patch is based on the correct version
		if patch.Version != 0 && patch.Version-1 != currentVersion {
			return fmt.Errorf("version mismatch: patch version %d, expected base %d", patch.Version, currentVersion)
		}

		newVersion := currentVersion + 1
		patch.Version = newVersion

		pipe := tx.TxPipeline()
		pipe.Set(c.Ctx, interview.VersionCacheKey, newVersion, interview.UntilExpire)

		patchJSON, err := json.Marshal(patch)
		if err != nil {
			return err
		}

		pipe.LPush(c.Ctx, interview.PatchKey, patchJSON)
		pipe.Expire(c.Ctx, interview.PatchKey, interview.UntilExpire)

		_, err = pipe.Exec(c.Ctx)
		if err != nil {
			return err
		}
		interview.Version = newVersion

		if newVersion%10 == 0 {
			go interview.CompactCodePatches()
		}

		return nil
	})
}
func (interview *Interview) CompactCodePatches() {
	c := interview.Cache

	code, err := interview.rebuildCodeFromPatches()
	if err != nil {
		return
	}
	pipe := c.Client.TxPipeline()

	state := CodeState{
		Content: code,
		Version: interview.Version,
	}
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return
	}

	pipe.Set(c.Ctx, interview.StateCacheKey, stateJSON, interview.UntilExpire)
	pipe.Del(c.Ctx, interview.PatchKey)

	_, err = pipe.Exec(c.Ctx)
	return
}

func (interview *Interview) rebuildCodeFromPatches() (string, error) {
	c := interview.Cache
	stateStr, err := c.Client.Get(c.Ctx, interview.StateCacheKey).Result()
	if err != nil {
		return "", err
	}
	var state CodeState
	var baseCode string
	if json.Unmarshal([]byte(stateStr), &state) == nil {
		baseCode = state.Content
	} else {
		return "", err
	}
	patchStrings, err := c.Client.LRange(c.Ctx, interview.PatchKey, 0, -1).Result()

	if err != nil || len(patchStrings) == 0 {
		return baseCode, nil
	}
	patches := make([]CodePatch, len(patchStrings))

	for _, patchStr := range patchStrings {
		var patch CodePatch
		if err = json.Unmarshal([]byte(patchStr), &patch); err != nil {
			continue
		}
		patches = append(patches, patch)
	}

	code := []rune(baseCode)
	slices.Reverse(patches)
	for _, patch := range patches {
		code = c.applyPatch(code, patch)
	}

	return string(code), nil
}

func (c *Cache) applyPatch(code []rune, patch CodePatch) []rune {
	codeLen := len(code)

	switch patch.Operation {
	case "add":
		if patch.StartPos < 0 {
			patch.StartPos = 0
		}
		if patch.StartPos > codeLen {
			patch.StartPos = codeLen
		}

		newContent := []rune(patch.Content)
		result := make([]rune, 0, codeLen+len(newContent))
		result = append(result, code[:patch.StartPos]...)
		result = append(result, newContent...)
		result = append(result, code[patch.StartPos:]...)
		return result

	case "remove":
		if patch.StartPos < 0 || patch.StartPos >= codeLen {
			return code
		}
		if patch.EndPos > codeLen {
			patch.EndPos = codeLen
		}
		if patch.EndPos <= patch.StartPos {
			return code
		}

		result := make([]rune, 0, codeLen-(patch.EndPos-patch.StartPos))
		result = append(result, code[:patch.StartPos]...)
		result = append(result, code[patch.EndPos:]...)
		return result

	case "replace":
		if patch.StartPos < 0 || patch.StartPos >= codeLen {
			return code
		}
		if patch.EndPos > codeLen {
			patch.EndPos = codeLen
		}
		if patch.EndPos < patch.StartPos {
			return code
		}

		newContent := []rune(patch.Content)
		removedLen := patch.EndPos - patch.StartPos
		result := make([]rune, 0, codeLen-removedLen+len(newContent))
		result = append(result, code[:patch.StartPos]...)
		result = append(result, newContent...)
		result = append(result, code[patch.EndPos:]...)
		return result
	}

	return code
}

func (interview *Interview) GetCurrentCode() (string, []CodePatch, int64, error) {
	c := interview.Cache

	stateStr, err := c.Client.Get(c.Ctx, interview.StateCacheKey).Result()
	if err == nil {
		var state CodeState
		if json.Unmarshal([]byte(stateStr), &state) == nil {
			if interview.Version == state.Version {
				return state.Content, []CodePatch{}, state.Version, nil
			}
		}
	}

	patchesArr, err := c.Client.LRange(c.Ctx, interview.PatchKey, 0, -1).Result()
	if err != nil {
		return "", []CodePatch{}, 0, err
	}

	var baseCode string
	var state CodeState
	if json.Unmarshal([]byte(stateStr), &state) == nil {
		baseCode = state.Content
	} else {
		return "", []CodePatch{}, 0, err
	}

	var codePatches []CodePatch
	for _, patchStr := range patchesArr {
		var patch CodePatch
		if json.Unmarshal([]byte(patchStr), &patch) == nil {
			codePatches = append(codePatches, patch)
		}
	}
	slices.Reverse(codePatches)
	return baseCode, codePatches, interview.Version, nil
}
