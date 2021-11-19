package object

import (
	"path"
	"tversion/pool"
	"tversion/utils"
)

// SaveObject 写入版本记录文件
//diff格式类似于
//{
//	"create":[
//		"aaa.txt",
//		"bbb.txt",
//	],
//	"create":[
//		"aaa.txt",
//		"bbb.txt",
//	],
//	"create":[
//		"aaa.txt",
//		"bbb.txt",
//	],
//}
func SaveObject(diff map[string][]string, oh *ObjectHistory, isfirst bool) {
	utils.DirCHeckAndMk(path.Join(DataPathDir, "objects"))
	p := pool.NewPool(1000)
	for _, cfilePath := range diff["create"] {
		p.Add(1)
		go func() {
			objectName := "version1"
			if !isfirst { //如果不是第一个版本则要记录下修改的内容，如果每一个版本都记录版本数据容易变大
				objectName = utils.StrToMd5(diff["version"][0] + cfilePath)
				objectPath := path.Join(DataPathDir, "objects", objectName)
				FileToBin(cfilePath, objectPath)
			}

			oh.AddObjHistory(cfilePath, diff["version"][0], "create", objectName)
			p.Done()
		}()
		p.Wait()
	}

	for _, ufilePath := range diff["update"] {
		p.Add(1)
		go func() {
			objectName := "version1"
			if !isfirst { //如果不是第一个版本则要记录下修改的内容，如果每一个版本都记录版本数据容易变大
				objectName = utils.StrToMd5(diff["version"][0] + ufilePath)
				objectPath := path.Join(DataPathDir, "objects", objectName)
				FileToBin(ufilePath, objectPath)
			}

			oh.AddObjHistory(ufilePath, diff["version"][0], "update", objectName)
			p.Done()
		}()
		p.Wait()
	}

	for _, dfilePath := range diff["delete"] {
		p.Add(1)
		go func() {
			objectName := ""
			if !isfirst { //如果不是第一个版本则要记录下修改的内容，如果每一个版本都记录版本数据容易变大
				objectName = utils.StrToMd5(diff["version"][0] + dfilePath)
				objectPath := path.Join(DataPathDir, "objects", objectName)
				FileToBin(dfilePath, objectPath)
				objectName = "version1"
			}

			oh.AddObjHistory(dfilePath, diff["version"][0], "delete", objectName)
			p.Done()
		}()
		p.Wait()
	}

	oh.WriteHistory()
}

func FileToBin(filePath string, binPath string) {
	utils.CopyFile(filePath, binPath)
}
