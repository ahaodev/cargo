package internal

import (
	"cargo/pkg"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

func today() string {
	return time.Now().Format("20060102")
}
func todayFileName() string {
	return today() + ".txt"
}

// 创建新文件并写入内容
func createNewFile(filePath string) error {
	// 创建新文件
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer file.Close()

	// 写入内容到文件
	_, err = file.WriteString("0")
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// ReadPassedCount 获取当前日期文件的数值
func ReadPassedCount() (int, error) {
	pkg.Log.Println("ReadPassedCount")
	// 如果文件不存在，则创建文件
	// 判断文件是否存在
	if _, err := os.Stat(todayFileName()); os.IsNotExist(err) {
		// 如果文件不存在，创建一个新文件并写入内容
		err := createNewFile(todayFileName())
		if err != nil {
			fmt.Println("创建文件失败:", err)
			return 0, nil
		}
		fmt.Println("新文件已创建")
	} else if err != nil {
		// 如果发生其他错误，输出错误信息
		fmt.Println("获取文件信息失败:", err)
		return 0, nil
	} else {
		fmt.Println("文件已存在")
	}
	// 读取文件内容
	content, err := os.ReadFile(todayFileName())
	if err != nil {
		fmt.Println(err)
		os.WriteFile(todayFileName(), []byte("0"), 0644)
		return 0, err
	}
	fmt.Println(content)
	// 将文件内容转换为整数
	number, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	return number, nil
}

// WritePassedCount 写入当前日期文件的数值
func WritePassedCount() (int, error) {
	number, err := ReadPassedCount()
	if err != nil {
		return 0, err
	}
	pkg.Log.Printf("Before number: %d\n", number)
	number += 1
	pkg.Log.Printf("After number: %d\n", number)
	// 将整数转换为字符串
	content := strconv.Itoa(number)
	pkg.Log.Println("WritePassedCount")
	// 写入文件内容
	err = ioutil.WriteFile(todayFileName(), []byte(content), 0644)
	return number, err
}

// ClearPassedCount 清除当前日期文件的数值
func ClearPassedCount() error {
	pkg.Log.Println("ClearPassedCount")
	// 如果文件不存在，则创建文件
	if _, err := os.Stat(todayFileName()); os.IsNotExist(err) {
		// 如果文件不存在，创建一个新文件并写入内容
		err := createNewFile(todayFileName())
		if err != nil {
			fmt.Println("创建文件失败:", err)
			return err
		}
		fmt.Println("新文件已创建")
	} else if err != nil {
		// 如果发生其他错误，输出错误信息
		fmt.Println("获取文件信息失败:", err)
		return err
	} else {
		fmt.Println("文件已存在")
	}
	return ioutil.WriteFile(todayFileName(), []byte("0"), 0644)
}
