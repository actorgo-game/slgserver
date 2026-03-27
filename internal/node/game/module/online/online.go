package online

import "sync"

type playerInfo struct {
	PlayerId  int64
	UID       int64
	AgentPath string
}

var (
	mu      sync.RWMutex
	players = make(map[int64]*playerInfo)   // uid -> info
	ridMap  = make(map[int]*playerInfo)      // rid -> info
)

func BindPlayer(rid int, uid int64, agentPath string) {
	mu.Lock()
	defer mu.Unlock()
	info := &playerInfo{PlayerId: int64(rid), UID: uid, AgentPath: agentPath}
	players[uid] = info
	ridMap[rid] = info
}

func UnBindPlayer(uid int64) {
	mu.Lock()
	defer mu.Unlock()
	if info, ok := players[uid]; ok {
		delete(ridMap, int(info.PlayerId))
		delete(players, uid)
	}
}

func GetAgentPath(rid int) string {
	mu.RLock()
	defer mu.RUnlock()
	if info, ok := ridMap[rid]; ok {
		return info.AgentPath
	}
	return ""
}

func Count() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(players)
}

func GetAllRIds() []int {
	mu.RLock()
	defer mu.RUnlock()
	rids := make([]int, 0, len(ridMap))
	for rid := range ridMap {
		rids = append(rids, rid)
	}
	return rids
}
