package facility

import (
	"encoding/json"
	"io/ioutil"
	"path"

	clog "github.com/actorgo-game/actorgo/logger"
	cprofile "github.com/actorgo-game/actorgo/profile"
)

const (
	Main          = 0  //主城
	JiaoChang     = 13 //校场
	TongShuaiTing = 14 //统帅厅
	JiShi         = 15 //集市
	MBS           = 16 //募兵所
)

var FConf facilityConf

type conf struct {
	Name string
	Type int8
}

type facilityConf struct {
	Title     string `json:"title"`
	List      []conf `json:"list"`
	facilitys map[int8]*facility
}

func (this *facilityConf) Load() {
	this.facilitys = make(map[int8]*facility, 0)

	jsonDir := cprofile.ConfigPath()
	fileName := path.Join(jsonDir, "facility", "facility.json")
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		clog.Panic("[facility] load file error: file=%s, err=%v", fileName, err)
		return
	}
	if err := json.Unmarshal(jdata, this); err != nil {
		clog.Panic("[facility] unmarshal error: file=%s, err=%v", fileName, err)
		return
	}

	fdir := path.Join(jsonDir, "facility")
	files, err := ioutil.ReadDir(fdir)
	if err != nil {
		return
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if file.Name() == "facility.json" || file.Name() == "facility_addition.json" {
			continue
		}
		fileName := path.Join(fdir, file.Name())
		f := NewFacility(fileName)
		this.facilitys[f.Type] = f
	}
	clog.Info("[facility] FConf loaded: %d facilities", len(this.facilitys))
}

func (this *facilityConf) MaxLevel(fType int8) int8 {
	f, ok := this.facilitys[fType]
	if ok {
		return int8(len(f.Levels))
	}
	return 0
}

func (this *facilityConf) Need(fType int8, level int8) (*NeedRes, bool) {
	if level <= 0 {
		return nil, false
	}
	f, ok := this.facilitys[fType]
	if ok {
		if int8(len(f.Levels)) >= level {
			return &f.Levels[level-1].Need, true
		}
		return nil, false
	}
	return nil, false
}

// 升级需要的时间
func (this *facilityConf) CostTime(fType int8, level int8) int {
	if level <= 0 {
		return 0
	}
	f, ok := this.facilitys[fType]
	if ok {
		if int8(len(f.Levels)) >= level {
			return f.Levels[level-1].Time - 2 //比客户端快2s，保证客户端倒计时完一定是升级成功了
		}
		return 0
	}
	return 0
}

func (this *facilityConf) GetValues(fType int8, level int8) []int {
	if level <= 0 {
		return []int{}
	}
	f, ok := this.facilitys[fType]
	if ok {
		if int8(len(f.Levels)) >= level {
			return f.Levels[level-1].Values
		}
		return []int{}
	}
	return []int{}
}

func (this *facilityConf) GetAdditions(fType int8) []int8 {
	f, ok := this.facilitys[fType]
	if ok {
		return f.Additions
	}
	return []int8{}
}
