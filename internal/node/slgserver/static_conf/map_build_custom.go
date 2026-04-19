package static_conf

import (
	clog "github.com/actorgo-game/actorgo/logger"

	"github.com/llr104/slgserver/internal/node/slgserver/static_conf/facility"
)

// 地图资源配置
var MapBCConf mapBuildCustomConf

type level struct {
	Level    int8             `json:"level"`
	Time     int              `json:"time"` //升级需要的时间
	Durable  int              `json:"durable"`
	Defender int              `json:"defender"`
	Need     facility.NeedRes `json:"need"`
	Result   result           `json:"result"`
}

type customConf struct {
	Type   int8    `json:"type"`
	Name   string  `json:"name"`
	Levels []level `json:"levels"`
}

type result struct {
	ArmyCnt int `json:"army_cnt"`
}

type BCLevelCfg struct {
	level
	Type int8   `json:"type"`
	Name string `json:"name"`
}

type mapBuildCustomConf struct {
	Title  string       `json:"title"`
	Cfg    []customConf `json:"cfg"`
	cfgMap map[int8]customConf
}

func (this *mapBuildCustomConf) Load() {
	LoadJSON("map_build_custom.json", this)

	this.cfgMap = make(map[int8]customConf)
	for _, v := range this.Cfg {
		this.cfgMap[v.Type] = v
	}
	clog.Info("[static_conf] map_build_custom loaded: %d types", len(this.cfgMap))
}

func (this *mapBuildCustomConf) BuildConfig(cfgType int8, level int8) (*BCLevelCfg, bool) {
	if c, ok := this.cfgMap[cfgType]; ok {
		if len(c.Levels) < int(level) {
			return nil, false
		}

		lc := c.Levels[level-1]
		cfg := BCLevelCfg{Type: cfgType, Name: c.Name}
		cfg.Level = level
		cfg.Need = lc.Need
		cfg.Result = lc.Result
		cfg.Durable = lc.Durable
		cfg.Time = lc.Time
		cfg.Result = lc.Result

		return &cfg, true
	}
	return nil, false
}

// 可容纳队伍数量
func (this *mapBuildCustomConf) GetHoldArmyCnt(cfgType int8, level int8) int {
	cfg, ok := this.BuildConfig(cfgType, level)
	if !ok {
		return 0
	}
	return cfg.Result.ArmyCnt
}
