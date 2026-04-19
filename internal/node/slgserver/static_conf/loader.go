package static_conf

import (
	"encoding/json"
	"io/ioutil"
	"path"

	clog "github.com/actorgo-game/actorgo/logger"
	cprofile "github.com/actorgo-game/actorgo/profile"
)

// JsonDir 返回 JSON 配置文件根目录。
// 与 cluster.json 中 "config_path" 一致；保持原 slgserver "../data/conf/" 的语义。
func JsonDir() string {
	return cprofile.ConfigPath()
}

// LoadJSON 从 JsonDir() 下的相对路径读取 JSON 文件并反序列化到 out。
// 例：LoadJSON("basic.json", &Basic) 或 LoadJSON("facility/facility.json", &this)
func LoadJSON(rel string, out any) {
	fileName := path.Join(JsonDir(), rel)
	jdata, err := ioutil.ReadFile(fileName)
	if err != nil {
		clog.Panic("[static_conf] load file error: file=%s, err=%v", fileName, err)
		return
	}
	if err := json.Unmarshal(jdata, out); err != nil {
		clog.Panic("[static_conf] unmarshal error: file=%s, err=%v", fileName, err)
		return
	}
}

// LoadFile 从 JsonDir() 下的相对路径读取文件原始字节。
func LoadFile(rel string) ([]byte, error) {
	fileName := path.Join(JsonDir(), rel)
	return ioutil.ReadFile(fileName)
}
