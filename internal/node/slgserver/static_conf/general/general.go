package general

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"path"
	"time"

	clog "github.com/actorgo-game/actorgo/logger"
	cprofile "github.com/actorgo-game/actorgo/profile"
)

var General general

type g struct {
	Name         string `json:"name"`
	CfgId        int    `json:"cfgId"`
	Force        int    `json:"force"`
	Strategy     int    `json:"strategy"`
	Defense      int    `json:"defense"`
	Speed        int    `json:"speed"`
	Destroy      int    `json:"destroy"`
	ForceGrow    int    `json:"force_grow"`
	StrategyGrow int    `json:"strategy_grow"`
	DefenseGrow  int    `json:"defense_grow"`
	SpeedGrow    int    `json:"speed_grow"`
	DestroyGrow  int    `json:"destroy_grow"`
	Cost         int8   `json:"cost"`
	Probability  int    `json:"probability"`
	Star         int8   `json:"star"`
	Arms         []int  `json:"arms"`
	Camp         int8   `json:"camp"`
}

type general struct {
	Title            string `json:"title"`
	GArr             []g    `json:"list"`
	GMap             map[int]g
	totalProbability int
}

func (this *general) Load() {
	fileName := path.Join(cprofile.ConfigPath(), "general", "general.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		clog.Panic("[general] load file error: file=%s, err=%v", fileName, err)
		return
	}

	this.totalProbability = 0
	if err := json.Unmarshal(jdata, this); err != nil {
		clog.Panic("[general] unmarshal error: file=%s, err=%v", fileName, err)
		return
	}
	this.GMap = make(map[int]g)
	for _, v := range this.GArr {
		this.GMap[v.CfgId] = v
		this.totalProbability += v.Probability
	}
	clog.Info("[general] General loaded: %d records", len(this.GArr))
}

func (this *general) Cost(cfgId int) int8 {
	c, ok := this.GMap[cfgId]
	if ok {
		return c.Cost
	}
	return 0
}

func (this *general) Draw() int {
	rand.Seed(time.Now().UnixNano())
	rate := rand.Intn(this.totalProbability)

	cur := 0
	for _, g := range this.GArr {
		if rate >= cur && rate < cur+g.Probability {
			return g.CfgId
		}
		cur += g.Probability
	}
	return 0
}
