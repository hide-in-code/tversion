package object

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"tversion/utils"
)

type History struct {
	Version    string `json:"version"`
	Action     string `json:"action"`
	ObjectName string `json:"objectName"`
}

type ObjectHistory struct {
	Files map[string][]History `json:"files"`
}

const (
	DataPathDir string = ".tversion/object"
)

func init() {
	utils.DirCHeckAndMk(path.Join(utils.GetWd(), DataPathDir))
}

func GetObjectHistory() *ObjectHistory {
	ohFilePath := path.Join(utils.GetWd(), DataPathDir, "object_history.json")
	_, err := os.Stat(ohFilePath)
	oh := &ObjectHistory{}
	oh.Files = map[string][]History{}

	if err == nil { //文件存在
		f, err := os.Open(ohFilePath)
		defer f.Close()
		if err != nil {
			panic(err)
		}
		r := io.Reader(f)

		if err = json.NewDecoder(r).Decode(oh); err != nil {
			panic(err)
		}
		return oh
	}

	oh.WriteHistory()
	return oh
}

func (oh *ObjectHistory) AddObjHistory(key string, version string, action string, objName string) {
	h := History{
		Version:    version,
		Action:     action,
		ObjectName: objName,
	}

	oh.Files[key] = append(oh.Files[key], h)
}

func (oh *ObjectHistory) WriteHistory() {
	ohFilePath := path.Join(utils.GetWd(), DataPathDir, "object_history.json")

	_, err := os.Stat(ohFilePath)
	if err == nil { //清除文件
		os.Remove(ohFilePath)
	}

	fp, err := os.OpenFile(ohFilePath, os.O_RDWR|os.O_CREATE, 0600)
	defer fp.Close()
	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(oh)
	if err != nil {
		panic(nil)
	}

	_, err = fp.Write(data)
	if err != nil {
		panic(nil)
	}
}
