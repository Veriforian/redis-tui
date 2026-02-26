package cmd

import (
	"github.com/davidbudnick/redis-tui/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

// Favorites

func LoadFavoritesCmd(connID int64) tea.Cmd {
	return func() tea.Msg {
		cfg := GetConfig()
		if cfg == nil {
			return types.FavoritesLoadedMsg{Err: nil}
		}
		favorites := cfg.ListFavorites(connID)
		return types.FavoritesLoadedMsg{Favorites: favorites, Err: nil}
	}
}

func AddFavoriteCmd(connID int64, key, label string) tea.Cmd {
	return func() tea.Msg {
		cfg := GetConfig()
		if cfg == nil {
			return types.FavoriteAddedMsg{Err: nil}
		}
		fav, err := cfg.AddFavorite(connID, key, label)
		return types.FavoriteAddedMsg{Favorite: fav, Err: err}
	}
}

func RemoveFavoriteCmd(connID int64, key string) tea.Cmd {
	return func() tea.Msg {
		cfg := GetConfig()
		if cfg == nil {
			return types.FavoriteRemovedMsg{Err: nil}
		}
		err := cfg.RemoveFavorite(connID, key)
		return types.FavoriteRemovedMsg{Key: key, Err: err}
	}
}

// Recent keys

func LoadRecentKeysCmd(connID int64) tea.Cmd {
	return func() tea.Msg {
		cfg := GetConfig()
		if cfg == nil {
			return types.RecentKeysLoadedMsg{Err: nil}
		}
		keys := cfg.ListRecentKeys(connID)
		return types.RecentKeysLoadedMsg{Keys: keys, Err: nil}
	}
}

func AddRecentKeyCmd(connID int64, key string, keyType types.KeyType) tea.Cmd {
	return func() tea.Msg {
		cfg := GetConfig()
		if cfg != nil {
			cfg.AddRecentKey(connID, key, keyType)
		}
		return nil
	}
}

// Templates

func LoadTemplatesCmd() tea.Cmd {
	return func() tea.Msg {
		cfg := GetConfig()
		if cfg == nil {
			return types.TemplatesLoadedMsg{Err: nil}
		}
		templates := cfg.ListTemplates()
		return types.TemplatesLoadedMsg{Templates: templates, Err: nil}
	}
}
