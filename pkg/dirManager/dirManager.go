package dirManager

import (
	"log"
	"os"
)

const (
	mddir   = "/home/user/programmin/obsidianTestEnv/data/obsidianProject/mddir"
	htmldir = "/home/user/programmin/obsidianTestEnv/data/obsidianProject/htmldir"
	pdfdir  = "/home/user/programmin/obsidianTestEnv/data/obsidianProject/pdfdir"
)

type DirManager struct {
	Alldir []string
}

func NewDirManager(dirrectories []string) *DirManager {
	return &DirManager{Alldir: dirrectories}
}

func (dm *DirManager) Check() {
	_, ok := dm.checkAllDirExists(dm.Alldir)

	if !ok {
		if err := dm.createDir(); err != nil {
			log.Println(err)
			return
		}
	} else {
		log.Println("all dir exists")
	}
}

func (dm *DirManager) checkDirExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false //dir does not exist
		}
		log.Println("error during check dir", err)
		return false

	}

	if b := info.IsDir(); b {
		return true
	}
	return false
}

func (dm *DirManager) checkAllDirExists(s []string) ([]string, bool) {
	b := true
	stack := make([]string, 0)
	for _, dir := range s {
		if exist := dm.checkDirExists(dir); !exist {
			b = false
			stack = append(stack, dir) // nonExisting dirs
		}
	}
	return stack, b
}

func (dm *DirManager) createDir() error {
	for _, path := range dm.Alldir {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
		log.Println("dir created", path)
	}
	return nil
}
