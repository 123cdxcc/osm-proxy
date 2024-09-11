package cache

import (
	"fmt"
	"io"
	"os"
)

var (
	homeDir = "cache"
)

func Init(dir string) error {
	homeDir = dir
	err := os.Mkdir(homeDir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("设置缓存: 创建目录: %w", err)
	}
	return nil
}

func Set(key string, r io.Reader) error {
	fs, err := os.OpenFile(fmt.Sprintf("%v/%v", homeDir, key), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("设置缓存: 创建文件: %w", err)
	}
	_, err = io.Copy(fs, r)
	if err != nil {
		return fmt.Errorf("设置缓存: 写入文件: %w", err)
	}
	return nil
}

func Get(key string) ([]byte, error) {
	fs, err := os.Open(fmt.Sprintf("%v/%v", homeDir, key))
	if err != nil {
		return nil, fmt.Errorf("获取缓存: 打开文件: %w", err)
	}
	b, err := io.ReadAll(fs)
	if err != nil {
		return nil, fmt.Errorf("获取缓存: 读取文件: %w", err)
	}
	return b, nil
}
