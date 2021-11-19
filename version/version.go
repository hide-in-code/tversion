package version

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
	"tversion/object"
	"tversion/pool"
	"tversion/utils"
)

const (
	dataPathDir string = ".tversion/version"
)

type Version struct {
	VersionNum string            `json:"versionNum"`
	Ver        map[string]string `json:"ver"`
}

func Commit(dir string, version string) {
	//历史版本检查
	vh := GetVersionHistory()
	if vh.HistoryExist(version) {
		fmt.Println("版本已经存在，不能commit到该版本", version)
		return
	}

	//上版本
	lastHistory, _ := vh.GetLastHistory()

	//版本整理
	totalFile := 0
	ver := make(map[string]string)
	files := utils.Walk(dir)
	p := pool.NewPool(10000)
	for _, filepath := range files {
		if strings.Index(filepath, ".tversion") != -1 {//跳过程序文件
			continue
		}

		p.Add(1)
		go func() {
			md5 := utils.GetMd5(filepath)
			ver[filepath] = md5
			totalFile += 1
			p.Done()
		}()
		p.Wait()
	}

	newVersion := &Version{
		VersionNum: version,
		Ver:        ver,
	}

	//版本写入 todo:后续改为只保存最后两个版
	vFilePath := path.Join(utils.GetWd(), dataPathDir, version+".json")
	fp, err := os.OpenFile(vFilePath, os.O_RDWR|os.O_CREATE, 0600)
	defer fp.Close()
	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(newVersion)
	if err != nil {
		panic(nil)
	}

	_, err = fp.Write(data)
	if err != nil {
		panic(nil)
	}

	//新增版本
	history := History{
		VersionNum: version,
		Time:       time.Now().Unix(),
	}
	vh.AddVersionHistory(history)

	//版本比对
	diffRes := diff(version, lastHistory.VersionNum)
	if len(diffRes["create"]) == 0 && len(diffRes["delete"]) == 0 && len(diffRes["update"]) == 0 {
		fmt.Println("版本未发生任何变化，版本构建放弃")
		deleteVersion(version)
		vh.DeleteVersionHistory(history)
		return //不再往后执行
	} else {
		for _, cfilePath := range diffRes["create"] {
			fmt.Println("新增文件", cfilePath)
		}
		for _, ufilePath := range diffRes["update"] {
			fmt.Println("修改文件", ufilePath)
		}
		for _, dfilePath := range diffRes["delete"] {
			fmt.Println("删除文件", dfilePath)
		}
		fmt.Println("新增文件：", len(diffRes["create"]))
		fmt.Println("修改文件：", len(diffRes["update"]))
		fmt.Println("删除文件：", len(diffRes["delete"]))
		fmt.Println("构建新版本", version)
	}

	//写入obj历史和每个变化的文件
	fmt.Println("正在写入object记录")
	oh := object.GetObjectHistory()

	firstVersion, err := vh.GetFirstHistory()
	if err != nil { //不存在历史版本，这种数据是不对的，程序执行到此处必定有至少一个版本
		panic(err)
	}

	//isfirst := false
	if firstVersion.VersionNum == diffRes["version"][0] {
		//isfirst = true
	}

	//object.SaveObject(diffRes, oh, isfirst)
	object.SaveObject(diffRes, oh, false)//暂时所有版本都保存
}

func deleteVersion(version string) {
	jsonPath := path.Join(utils.GetWd(), dataPathDir, version+".json")
	_, err := os.Stat(jsonPath)
	if err == nil { //文件存在
		os.Remove(jsonPath)
	}
}

func getVersionData(version string) *Version {
	jsonPath := path.Join(utils.GetWd(), dataPathDir, version+".json")
	f, err := os.Open(jsonPath)
	defer f.Close()
	r := io.Reader(f)
	ret := &Version{}

	if err != nil {
		return ret
	}

	if err = json.NewDecoder(r).Decode(ret); err != nil {
		panic(err)
	}
	return ret
}

func keyExist(key string, m map[string]string) bool {
	if _, ok := m[key]; ok {
		return true
	}

	return false
}

func diffDelete(bar *Version, source *Version) map[string][]string {
	barMap := bar.Ver
	sourceMap := source.Ver

	p := pool.NewPool(100)
	ret := make(map[string][]string)
	ret["delete"] = []string{}

	for path, _ := range sourceMap {
		p.Add(1)
		go func() {
			if !keyExist(path, barMap) { //新版本中找不到旧版文件路径即为删除
				ret["delete"] = append(ret["delete"], path)
				p.Done()
				return
			}
			p.Done()
		}()
		p.Wait()
	}

	return ret
}

func diffAddUpdate(bar *Version, source *Version) map[string][]string {
	barMap := bar.Ver
	sourceMap := source.Ver

	ret := make(map[string][]string)
	ret["create"] = []string{}
	ret["update"] = []string{}

	p := pool.NewPool(100)
	for path, md5 := range barMap {
		p.Add(1)
		go func() {
			if !keyExist(path, sourceMap) {
				ret["create"] = append(ret["create"], path)
				p.Done()
				return
			}

			if md5 != sourceMap[path] {
				ret["update"] = append(ret["update"], path)
				//fmt.Println("=================")
				//fmt.Println("文件发生修改")
				//fmt.Println("当前md5", md5)
				//fmt.Println("原始md5", sourceMap[path])
				//fmt.Println("文件路径", path)
				//fmt.Println("=================")
			}

			p.Done()
		}()
		p.Wait()
	}

	return ret
}

func diff(version1 string, version2 string) map[string][]string {
	ret := make(map[string][]string)
	ret["version"] = []string{version1}

	v1Data := getVersionData(version1)
	v2Data := getVersionData(version2)
	diff1 := diffAddUpdate(v1Data, v2Data)
	diff2 := diffDelete(v1Data, v2Data)

	ret["create"] = diff1["create"]
	ret["update"] = diff1["update"]
	ret["delete"] = diff2["delete"]

	return ret
}


func Checkout(versionNum string)  {
	vh := GetVersionHistory()
	vhSlice := []string{}
	verCheckRight := false
	for _, h := range vh.History {
		if !verCheckRight { //将第一版本到回退到的版本全部放到切片，供后面每个文件计算使用
			vhSlice = append(vhSlice, h.VersionNum)
			if h.VersionNum == versionNum {
				verCheckRight = true
			}
		}
	}

	oh := object.GetObjectHistory()
	for key, versions := range oh.Files {//每个文件的所有历史版本
		lastVersion := versions[len(versions) - 1]

		handleObjectName := ""
		handleVersionName := ""
		handleAction := ""
		for _, oVer := range versions {//找到最后一次版本
			for _, hVer := range vhSlice {
				if oVer.Version == hVer {
					handleObjectName = oVer.ObjectName
					handleVersionName = oVer.Version
					handleAction = oVer.Action
				}
			}
		}

		//未匹配到任何版本，该文件还未进入版本库，应该做文件删除
		if handleObjectName == "" && handleVersionName == "" && handleAction == "" {
			_, err := os.Stat(key)
			if err == nil { //清除文件
				os.Remove(key)
			}

			fmt.Println("正在恢复文件", key, "从", lastVersion.Version, "到【删除】")
			continue
		}

		//最后版本是删除操作
		if handleAction == "delete" {
			_, err := os.Stat(key)
			if err == nil { //清除文件
				os.Remove(key)
			}

			fmt.Println("正在恢复文件", key, "从", lastVersion.Version, "到【删除】")
			continue
		}

		//如果文件最后版本和需要恢复版本是同一版本，则无需处理
		if lastVersion.Version == handleVersionName {
			fmt.Println("无须处理", key)
			continue
		}


		fmt.Println("正在恢复文件", key, "从", lastVersion.Version, "到", handleVersionName)
		objectPath := path.Join(object.DataPathDir, "objects", handleObjectName)
		object.FileToBin(objectPath, key)
	}
}


////版本切换
//func Checkout1(versionNum string)  {
//	vh := GetVersionHistory()
//	vhSlice := map[int]string{}
//	verCheckRight := false
//	revertIndex := 0
//	for index, h := range vh.History {
//		vhSlice[index] = h.VersionNum
//		if h.VersionNum == versionNum {
//			verCheckRight = true
//			revertIndex = index
//		}
//	}
//
//	if !verCheckRight {
//		fmt.Println("不存在的版本", versionNum)
//		return
//	}
//
//	//一个文件应该回到对应的版本
//	oh := object.GetObjectHistory()
//	for key, versions := range oh.Files {
//
//		//文件归属版本算法
//		lastVersion := versions[len(versions) - 1]
//		handleVersionSlice := versions[0:revertIndex + 1]
//
//		handleObjectName := ""
//		handleVersionName := ""
//		handleAction := ""
//		for _, ver := range handleVersionSlice {
//			if ver.Version == "" {
//				continue
//			}
//			handleVersionName = ver.Version
//			handleObjectName = ver.ObjectName
//			handleAction = ver.Action
//		}
//
//
//		if handleAction == "delete" {
//			_, err := os.Stat(key)
//			if err == nil { //清除文件
//				os.Remove(key)
//			}
//
//			fmt.Println("正在恢复文件", key, "cong", lastVersion.Version, "到【删除】")
//			continue
//		}
//
//
//		if handleObjectName != "version1" {//第一次提交的文件不需要处理
//			if handleAction == "delete" { //delete清除文件
//				_, err := os.Stat(key)
//				if err == nil { //清除文件
//					os.Remove(key)
//				}
//			} else {
//				fmt.Println("正在恢复文件", key, "cong", lastVersion.Version, "到", handleVersionName)
//				objectPath := path.Join(object.DataPathDir, "objects", handleObjectName)
//				object.FileToBin(objectPath, key)
//
//				//todo 该文件的历史版本回退？
//			}
//		}
//
//
//	}
//}
