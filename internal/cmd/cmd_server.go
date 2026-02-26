package cmd

import (
	"github.com/davidbudnick/redis-tui/internal/types"

	tea "github.com/charmbracelet/bubbletea"
)

func LoadServerInfoCmd() tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.ServerInfoLoadedMsg{Err: nil}
		}
		info, err := rc.GetServerInfo()
		return types.ServerInfoLoadedMsg{Info: info, Err: err}
	}
}

func FlushDBCmd() tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.FlushDBMsg{Err: nil}
		}
		err := rc.FlushDB()
		return types.FlushDBMsg{Err: err}
	}
}

func SwitchDBCmd(dbNum int) tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.DBSwitchedMsg{DB: dbNum, Err: nil}
		}
		err := rc.SelectDB(dbNum)
		return types.DBSwitchedMsg{DB: dbNum, Err: err}
	}
}

func GetSlowLogCmd(count int64) tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.SlowLogLoadedMsg{Err: nil}
		}
		entries, err := rc.SlowLogGet(count)
		return types.SlowLogLoadedMsg{Entries: entries, Err: err}
	}
}

func EvalLuaScriptCmd(script string, keys []string, args ...interface{}) tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.LuaScriptResultMsg{Err: nil}
		}
		result, err := rc.Eval(script, keys, args...)
		return types.LuaScriptResultMsg{Result: result, Err: err}
	}
}

func PublishMessageCmd(channel, message string) tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.PublishResultMsg{Channel: channel, Err: nil}
		}
		receivers, err := rc.Publish(channel, message)
		return types.PublishResultMsg{Channel: channel, Receivers: receivers, Err: err}
	}
}

func GetPubSubChannelsCmd(pattern string) tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.PubSubChannelsLoadedMsg{Err: nil}
		}
		names, err := rc.PubSubChannels(pattern)
		if err != nil {
			return types.PubSubChannelsLoadedMsg{Err: err}
		}
		channels := make([]types.PubSubChannel, len(names))
		for i, name := range names {
			channels[i] = types.PubSubChannel{Name: name}
		}
		return types.PubSubChannelsLoadedMsg{Channels: channels}
	}
}

func GetClientListCmd() tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.ClientListLoadedMsg{Err: nil}
		}
		clients, err := rc.ClientList()
		return types.ClientListLoadedMsg{Clients: clients, Err: err}
	}
}

func GetMemoryStatsCmd() tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.MemoryStatsLoadedMsg{Err: nil}
		}
		stats, err := rc.GetMemoryStats()
		return types.MemoryStatsLoadedMsg{Stats: stats, Err: err}
	}
}

func GetClusterInfoCmd() tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.ClusterInfoLoadedMsg{Err: nil}
		}
		nodes, err := rc.ClusterNodes()
		info, _ := rc.ClusterInfo()
		return types.ClusterInfoLoadedMsg{Nodes: nodes, Info: info, Err: err}
	}
}

func FetchClusterNodesCmd() tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.ClusterNodesLoadedMsg{Err: nil}
		}
		nodes, err := rc.ClusterNodes()
		return types.ClusterNodesLoadedMsg{Nodes: nodes, Err: err}
	}
}

func LoadRedisConfigCmd(pattern string) tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.ConfigLoadedMsg{Err: nil}
		}
		params, err := rc.ConfigGet(pattern)
		return types.ConfigLoadedMsg{Params: params, Err: err}
	}
}

func SetRedisConfigCmd(param, value string) tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.ConfigSetMsg{Param: param, Value: value, Err: nil}
		}
		err := rc.ConfigSet(param, value)
		return types.ConfigSetMsg{Param: param, Value: value, Err: err}
	}
}

func LoadLiveMetricsCmd() tea.Cmd {
	return func() tea.Msg {
		rc := getRedisClient()
		if rc == nil {
			return types.LiveMetricsMsg{Err: nil}
		}
		data, err := rc.GetLiveMetrics()
		return types.LiveMetricsMsg{Data: data, Err: err}
	}
}
