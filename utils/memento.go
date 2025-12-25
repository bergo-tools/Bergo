package utils

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
)

/*
memento文件，跟随session，用于存储agent的状态
agent执行task时将对应session的memento文件从.bergo目录下移到项目根目录，使得可以被git跟踪到
agent执行task完成后，将memento文件从项目根目录移回.bergo目录
时间线回退后，将回退的memento文件从根目录移到.bergo目录
*/
func InitMementoFile(sessionID string) {
	bergoFileName := sessionID + ".bergo.memento"
	mementoFilePath := filepath.Join(GetWorkspaceStorePath(), bergoFileName)
	if _, err := os.Stat(mementoFilePath); err != nil {
		if os.IsNotExist(err) {
			os.WriteFile(mementoFilePath, []byte(""), 0644)
		} else {
			panic(err)
		}
	}
	mementoContent, err := os.ReadFile(mementoFilePath)
	if err != nil {
		panic(err)
	}
	os.WriteFile("./.bergo.memento", mementoContent, 0644)

}

func HideMementoFile(sessionID string) {
	mementoContent, err := os.ReadFile("./.bergo.memento")
	if err != nil {
		panic(err)
	}
	bergoFileName := sessionID + ".bergo.memento"
	mementoFilePath := filepath.Join(GetWorkspaceStorePath(), bergoFileName)
	err = os.WriteFile(mementoFilePath, mementoContent, 0644)
	if err != nil {
		panic(err)
	}
	os.Remove("./.bergo.memento")
}

// GetMementoHash 获取当前 memento 文件的 MD5 哈希值
func GetMementoHash() string {
	content, err := os.ReadFile("./.bergo.memento")
	if err != nil {
		return ""
	}
	hash := md5.Sum(content)
	return hex.EncodeToString(hash[:])
}

// IsMementoChanged 检查 memento 文件是否有改动
func IsMementoChanged(initialHash string) bool {
	currentHash := GetMementoHash()
	return currentHash != initialHash
}
