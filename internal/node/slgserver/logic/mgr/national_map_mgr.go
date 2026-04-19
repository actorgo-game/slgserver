package mgr

import (
	"encoding/json"
	"os"
	"path"
	"sync"

	clog "github.com/actorgo-game/actorgo/logger"
	cprofile "github.com/actorgo-game/actorgo/profile"

	"github.com/llr104/slgserver/internal/data/model"
	"github.com/llr104/slgserver/internal/node/slgserver/global"
	"github.com/llr104/slgserver/internal/util"
)

// mapData 与原 slgserver national_map_mgr.go 中 mapData 同结构。
type mapData struct {
	Width  int     `json:"w"`
	Height int     `json:"h"`
	List   [][]int `json:"list"`
}

type NationalMapMgr struct {
	mutex    sync.RWMutex
	conf     map[int]model.NationalMap
	sysBuild map[int]model.NationalMap
}

var NMMgr = &NationalMapMgr{
	conf:     make(map[int]model.NationalMap),
	sysBuild: make(map[int]model.NationalMap),
}

// Load 从 cprofile.ConfigPath()/map.json 读取整张大地图。
// 与原 slgserver 同语义：根据 width/height 把扁平 list 转成 (x, y, type, level) 网格。
func (this *NationalMapMgr) Load() {
	fileName := path.Join(cprofile.ConfigPath(), "map.json")
	jdata, err := os.ReadFile(fileName)
	if err != nil {
		clog.Error("[NMMgr] read map file err=%v file=%s", err, fileName)
		return
	}

	m := &mapData{}
	if err := json.Unmarshal(jdata, m); err != nil {
		clog.Error("[NMMgr] unmarshal map err=%v", err)
		return
	}

	global.MapWith = m.Width
	global.MapHeight = m.Height
	model.MapBound.W = m.Width
	model.MapBound.H = m.Height

	for i, v := range m.List {
		t := int8(v[0])
		l := int8(v[1])
		d := model.NationalMap{
			Y: i / global.MapHeight, X: i % global.MapWith,
			MId: i, Type: t, Level: l,
		}
		this.conf[i] = d
		if d.Type == model.MapBuildSysCity || d.Type == model.MapBuildSysFortress {
			this.sysBuild[i] = d
		}
	}
	clog.Info("[NMMgr] map loaded: w=%d h=%d sysBuild=%d",
		global.MapWith, global.MapHeight, len(this.sysBuild))
}

func (this *NationalMapMgr) IsCanBuild(x, y int) bool {
	posIndex := global.ToPosition(x, y)
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	c, ok := this.conf[posIndex]
	if !ok {
		return false
	}
	return c.Type != 0
}

// IsCanBuildCity 与原 slgserver 同语义：系统城池 5 格内 + 玩家建筑/城池占据 不可建。
func (this *NationalMapMgr) IsCanBuildCity(x, y int) bool {
	for _, n := range this.sysBuild {
		if n.Type == model.MapBuildSysCity {
			if x >= n.X-5 && x <= n.X+5 && y >= n.Y-5 && y <= n.Y+5 {
				return false
			}
		}
	}

	for i := x - 2; i <= x+2; i++ {
		if i < 0 || i > global.MapWith {
			return false
		}
		for j := y - 2; j <= y+2; j++ {
			if j < 0 || j > global.MapHeight {
				return false
			}
		}
		if !this.IsCanBuild(x, y) || !RBMgr.IsEmpty(x, y) || !RCMgr.IsEmpty(x, y) {
			return false
		}
	}
	return true
}

func (this *NationalMapMgr) MapResTypeLevel(x, y int) (bool, int8, int8) {
	if n, ok := this.PositionBuild(x, y); ok {
		return true, n.Type, n.Level
	}
	return false, 0, 0
}

func (this *NationalMapMgr) PositionBuild(x, y int) (model.NationalMap, bool) {
	posIndex := global.ToPosition(x, y)
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	b, ok := this.conf[posIndex]
	return b, ok
}

func (this *NationalMapMgr) Scan(x, y int) []model.NationalMap {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	minX := util.MaxInt(0, x-ScanWith)
	maxX := util.MinInt(40, x+ScanWith)
	minY := util.MaxInt(0, y-ScanHeight)
	maxY := util.MinInt(40, y+ScanHeight)

	c := (maxX - minX + 1) * (maxY - minY + 1)
	r := make([]model.NationalMap, c)
	index := 0
	for i := minX; i <= maxX; i++ {
		for j := minY; j <= maxY; j++ {
			if v, ok := this.conf[global.ToPosition(i, j)]; ok {
				r[index] = v
			}
			index++
		}
	}
	return r
}

// SysBuilds 用于 RBMgr.Load 时初始化系统建筑。
func (this *NationalMapMgr) SysBuilds() map[int]model.NationalMap {
	return this.sysBuild
}
