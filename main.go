package main

import (
	"fmt"
	"os"
	"time"
	"tversion/utils"
	"tversion/version"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法", "tver commit|versions|checkout")
		os.Exit(1)
	}

	if os.Args[1] != "checkout" && os.Args[1] != "commit" && os.Args[1] != "versions" {
		fmt.Println("用法", "tver commit|versions|checkout")
		os.Exit(1)
	}

	if os.Args[1] == "checkout" && len(os.Args) < 3 {
		fmt.Println("必须输入要checkout到的版本")
		os.Exit(1)
	}

	action := os.Args[1]
	versionNum := ""
	if os.Args[1] == "checkout" {
		versionNum = os.Args[2]
	}

	start := time.Now()
	if action == "commit" {
		version.Commit(utils.GetWd(), utils.Tmd5())
		elapsed := time.Since(start)
		fmt.Printf("执行时长 %s", elapsed)
		os.Exit(0)
	}

	if action == "versions" {
		if len(os.Args) < 3 {
			version.ShowVersions()
			elapsed := time.Since(start)
			fmt.Printf("执行时长 %s", elapsed)
			os.Exit(0)
		} else {
			version.ShowFileVersion(os.Args[2])
			elapsed := time.Since(start)
			fmt.Printf("执行时长 %s", elapsed)
			os.Exit(0)
		}
	}

	if action == "checkout" {
		version.Checkout(versionNum)
		elapsed := time.Since(start)
		fmt.Printf("执行时长 %s", elapsed)
		os.Exit(0)
	}
}
