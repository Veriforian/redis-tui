package cmd

import (
	"github.com/davidbudnick/redis-tui/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

func SearchByValueCmd(pattern, valueSearch string, maxKeys int) tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.KeysLoadedMsg{Err: nil}
		}
		keys, err := rc.SearchByValue(pattern, valueSearch, maxKeys)
		return types.KeysLoadedMsg{Keys: keys, Cursor: 0, Err: err}
	}
}

func RegexSearchCmd(pattern string, maxKeys int) tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.RegexSearchResultMsg{Err: nil}
		}
		keys, err := rc.ScanKeysWithRegex(pattern, maxKeys)
		return types.RegexSearchResultMsg{Keys: keys, Err: err}
	}
}

func FuzzySearchCmd(term string, maxKeys int) tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.FuzzySearchResultMsg{Err: nil}
		}
		keys, err := rc.FuzzySearchKeys(term, maxKeys)
		return types.FuzzySearchResultMsg{Keys: keys, Err: err}
	}
}

func CompareKeysCmd(key1, key2 string) tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.CompareKeysResultMsg{Err: nil}
		}
		val1, val2, err := rc.CompareKeys(key1, key2)
		return types.CompareKeysResultMsg{Key1Value: val1, Key2Value: val2, Err: err}
	}
}
