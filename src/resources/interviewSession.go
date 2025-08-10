package resources

import (
	"CodeStream/src"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type Interview struct {
	SessionID       string
	Language        string
	Version         int64
	StateCacheKey   string
	VersionCacheKey string
	PatchKey        string
	LanguageKey     string
	Cache           *Cache
}

type CodePatch struct {
	Version   int64  `json:"version"`
	Operation string `json:"op"`
	StartPos  int    `json:"start_pos"`
	EndPos    int    `json:"end_pos"`
	Content   string `json:"content"`
}

type CodeState struct {
	Content string `json:"content"`
	Version int64  `json:"version"`
}

func CreateInterviewSession(c *Cache, sessionID string) (Interview, error, bool) {
	defaultLanguage := src.Config.Languages[0]
	stateKey := fmt.Sprintf("session:%s:state", sessionID)
	currentLanguageKey := fmt.Sprintf("session:%s:lang", sessionID)
	versionKey := fmt.Sprintf("session:%s:version", sessionID)
	patchKey := fmt.Sprintf("session:%s:patch", sessionID)

	exists, err := c.Client.Exists(c.Ctx, stateKey).Result()
	if err != nil {
		return Interview{}, err, false
	}
	if exists > 0 {
		existInterview, err := GetInterviewSession(c, sessionID)
		if err != nil {
			return Interview{}, err, false
		}
		return existInterview, nil, false
	}

	state := CodeState{
		Content: "",
		Version: 1,
	}

	stateJSON, err := json.Marshal(state)
	if err != nil {
		return Interview{}, err, false
	}

	pipe := c.Client.TxPipeline()
	pipe.Set(c.Ctx, stateKey, stateJSON, time.Hour*24)
	pipe.Set(c.Ctx, versionKey, state.Version, time.Hour*24)
	pipe.Set(c.Ctx, currentLanguageKey, defaultLanguage, time.Hour*24)
	pipe.LPush(c.Ctx, patchKey, "", time.Hour*24)
	pipe.LTrim(c.Ctx, patchKey, 1, 0)
	_, err = pipe.Exec(c.Ctx)
	return Interview{
		SessionID:       sessionID,
		Language:        defaultLanguage,
		Version:         state.Version,
		Cache:           c,
		StateCacheKey:   stateKey,
		PatchKey:        patchKey,
		VersionCacheKey: versionKey,
		LanguageKey:     currentLanguageKey,
	}, nil, true
}

func (interview *Interview) EditLanguage(newLang string) error {
	for _, lang := range src.Config.Languages {
		if lang == newLang {
			interview.Cache.Set(interview.LanguageKey, newLang, time.Hour*24)
			interview.Language = lang
			return nil
		}
	}
	return fmt.Errorf("not found language")
}

func GetInterviewSession(c *Cache, sessionID string) (Interview, error) {
	stateKey := fmt.Sprintf("session:%s:state", sessionID)
	currentLanguageKey := fmt.Sprintf("session:%s:lang", sessionID)
	versionKey := fmt.Sprintf("session:%s:version", sessionID)
	patchKey := fmt.Sprintf("session:%s:patch", sessionID)

	langVal := c.Get(currentLanguageKey)
	if langVal == nil {
		return Interview{}, fmt.Errorf("failed to get language: %w")
	}
	language, ok := langVal.(string)
	if !ok {
		return Interview{}, fmt.Errorf("language is not a string")
	}

	versionVal := c.Get(versionKey)
	if versionVal == nil {
		return Interview{}, fmt.Errorf("failed to get version: %w")
	}
	versionStr, ok := versionVal.(string)
	if !ok {
		return Interview{}, fmt.Errorf("version is not a string")
	}
	version, err := strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		return Interview{}, fmt.Errorf("invalid version format: %w", err)
	}

	return Interview{
		SessionID:       sessionID,
		Language:        language,
		Version:         version,
		StateCacheKey:   stateKey,
		VersionCacheKey: versionKey,
		PatchKey:        patchKey,
		LanguageKey:     currentLanguageKey,
		Cache:           c,
	}, nil
}
