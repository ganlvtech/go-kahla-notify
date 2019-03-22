package main

import (
	"fmt"
	"github.com/ganlvtech/go-kahla-notify/kahla"
	"io/ioutil"
	"os"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func GetHeadImgFilePathWithCache(client *kahla.Client, headImgFileKey int, cachePath string) (string, error) {
	path := fmt.Sprintf("%s%c%d.png", cachePath, os.PathSeparator, headImgFileKey)
	if fileExists(path) {
		return path, nil
	}
	data, err := client.Oss.HeadImgFile(headImgFileKey)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return "", err
	}
	return path, nil
}
