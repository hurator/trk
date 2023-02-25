package files

import (
	"embed"
	"fmt"
	"os"
	"path"
)

//go:embed images/*.png
var Images embed.FS

var tmpFiles map[string]string

func init() {
	tmpFiles = make(map[string]string)
}

func CleanUp() {
	for _, val := range tmpFiles {
		err := os.Remove(val)
		if err != nil {
			fmt.Println("cleanup failed", val, err.Error())
		}
	}
}

func GetIconPath(name string) string {
	if p, ok := tmpFiles[name]; ok {
		return p
	}
	data, err := Images.ReadFile(name)
	if err != nil {
		panic(err)
	}
	file, err := os.CreateTemp("", "icon*"+path.Ext(name))
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		panic(err)
	}
	tmpFiles[name] = file.Name()
	return tmpFiles[name]
}
