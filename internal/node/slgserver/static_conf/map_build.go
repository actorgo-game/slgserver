package static_conf

import (
	clog "github.com/actorgo-game/actorgo/logger"
)

// 地图资源配置
var MapBuildConf mapBuildConf

type cfg struct {
	Type     int8   `json:"type"`
	Name     string `json:"name"`
	Level    int8   `json:"level"`
	Grain    int    `json:"grain"`
	Wood     int    `json:"wood"`
	Iron     int    `json:"iron"`
	Stone    int    `json:"stone"`
	Durable  int    `json:"durable"`
	Defender int    `json:"defender"`
}

type mapBuildConf struct {
	Title  string `json:"title"`
	Cfg    []cfg  `json:"cfg"`
	cfgMap map[int8][]cfg
}

func (this *mapBuildConf) Load() {
	LoadJSON("map_build.json", this)

	this.cfgMap = make(map[int8][]cfg)
	for _, v := range this.Cfg {
		if _, ok := this.cfgMap[v.Type]; !ok {
			this.cfgMap[v.Type] = make([]cfg, 0)
		}
		this.cfgMap[v.Type] = append(this.cfgMap[v.Type], v)
	}
	clog.Info("[static_conf] map_build loaded: %d types", len(this.cfgMap))
}

func (this *mapBuildConf) BuildConfig(cfgType int8, level int8) (*cfg, bool) {
	if c, ok := this.cfgMap[cfgType]; ok {
		for _, v := range c {
			if v.Level == level {
				return &v, true
			}
		}
	}
	return nil, false
}
