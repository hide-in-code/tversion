package version

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"time"
	"tversion/object"
	"tversion/utils"
)

type History struct {
	VersionNum string `json:"version_num"`
	Time       int64  `json:"time"`
}

type VersionHistory struct {
	History []History `json:"history"`
}

func init() {
	utils.DirCHeckAndMk(path.Join(utils.GetWd(), dataPathDir))
}

func (vh *VersionHistory) writeHistory() {
	vhFilePath := path.Join(utils.GetWd(), dataPathDir, "version_history.json")

	_, err := os.Stat(vhFilePath)
	if err == nil { //清除文件
		os.Remove(vhFilePath)
	}

	fp, err := os.OpenFile(vhFilePath, os.O_RDWR|os.O_CREATE, 0600)
	defer fp.Close()
	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(vh)
	if err != nil {
		panic(nil)
	}

	_, err = fp.Write(data)
	if err != nil {
		panic(nil)
	}
}

func GetVersionHistory() *VersionHistory {
	vhFilePath := path.Join(utils.GetWd(), dataPathDir, "version_history.json")
	_, err := os.Stat(vhFilePath)
	vh := &VersionHistory{}
	if err == nil { //文件存在
		f, err := os.Open(vhFilePath)
		defer f.Close()
		if err != nil {
			panic(err)
		}
		r := io.Reader(f)

		if err = json.NewDecoder(r).Decode(vh); err != nil {
			panic(err)
		}
		return vh
	}

	vh.writeHistory()
	return vh
}

func (vh *VersionHistory) AddVersionHistory(h History) {
	vh.History = append(vh.History, h)
	vh.writeHistory()
}

func (vh *VersionHistory) DeleteVersionHistory(h History) {
	history := []History{}
	for _, ih := range vh.History {
		if ih.VersionNum != h.VersionNum {
			history = append(history, ih)
		}
	}
	vh.History = history
	vh.writeHistory()
}

func (vh *VersionHistory) HistoryExist(version string) bool {
	for _, h := range vh.History {
		if h.VersionNum == version {
			return true
		}
	}

	return false
}

//获取第一个版本
func (vh *VersionHistory) GetFirstHistory() (History, error) {
	length := len(vh.History)
	if length == 0 {
		var err error = errors.New("no history")
		return History{}, err
	}

	return vh.History[0], nil
}

//获取最后一个版本
func (vh *VersionHistory) GetLastHistory() (History, error) {
	length := len(vh.History)
	if length == 0 {
		var err error = errors.New("no history")
		return History{}, err
	}

	return vh.History[length-1], nil
}

//获取所有版本
func ShowVersions()  {
	vh := GetVersionHistory()
	for _, h := range vh.History {
		timeobj := time.Unix(int64(h.Time), 0)
		date := timeobj.Format("2006-01-02 15:04:05")
		fmt.Println(h.VersionNum, date)
	}
}

func ShowFileVersion(key string)  {
	//整理版本时间
	timeMap := map[string]string{}
	vh := GetVersionHistory()
	for _, h := range vh.History {
		timeobj := time.Unix(int64(h.Time), 0)
		date := timeobj.Format("2006-01-02 15:04:05")
		timeMap[h.VersionNum] = date
	}


	oh := object.GetObjectHistory()
	if _, ok := oh.Files[key]; !ok {
		fmt.Println("未查到任何提交历史", key)
		return
	}

	foh := oh.Files[key]
	for _, h := range foh {
		fdate := timeMap[h.Version]
		fmt.Println(fdate, h.Version, h.Action, h.ObjectName)
	}
}
