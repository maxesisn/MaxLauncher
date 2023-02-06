package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"epminecraft-go/auth"
	"epminecraft-go/exit"
	"epminecraft-go/logger"
	"golang.org/x/crypto/salsa20"
	"io"
	"os"
	"path/filepath"
)

var log = logger.Logger()

var key [32]byte
var keyStr = "Ellye i loveth theei loveth thee"

func LoadNonce() []byte {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Error("无法获取用户主目录。")
		exit.LauncherExit(err)
	}
	nonce := make([]byte, 24)
	nonceFile, err := os.Open(dirname + "/.config/MaxLauncherData/IV.dat")
	if err != nil {
		_, err = rand.Read(nonce)
		if err != nil {
			log.Error("随机数生成失败。")
			exit.LauncherExit(err)
		}
		_, err = nonceFile.Write(nonce)
		if err != nil {
			log.Error("无法写入随机数。")
			exit.LauncherExit(err)
		}
	} else {
		_, err = nonceFile.Read(nonce[:])
		if err != nil {
			log.Error("无法读取随机数。")
			exit.LauncherExit(err)
		}
		_ = nonceFile.Close()
	}
	return nonce
}

func SaveAccount(profile auth.Profile) {
	var profileBytes bytes.Buffer
	encoder := gob.NewEncoder(&profileBytes)
	err := encoder.Encode(profile)
	if err != nil {
		log.Error("编码账户信息时出现错误。")
		exit.LauncherExit(err)
	}
	copy(key[:], keyStr)
	nonce := LoadNonce()
	salsa20.XORKeyStream(profileBytes.Bytes(), profileBytes.Bytes(), nonce, &key)
	ex, err := os.Executable()
	if err != nil {
		log.Error("无法获取可执行文件路径。")
		exit.LauncherExit(err)
	}
	dirname := filepath.Dir(ex)
	dirname = filepath.Join(dirname, "launcher_data")
	err = os.MkdirAll(dirname, os.ModePerm)
	if err != nil {
		log.Error("无法创建数据目录。")
		exit.LauncherExit(err)
	}
	profileFile, err := os.OpenFile(dirname+"/account.enc", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Error("无法操作账户数据文件。")
		exit.LauncherExit(err)
	}
	_, err = profileFile.Write(profileBytes.Bytes())
	if err != nil {
		log.Error("写入账户数据时失败。")
		exit.LauncherExit(err)
	}
	_ = profileFile.Close()
}

func LoadAccount() auth.Profile {
	var profile auth.Profile
	ex, err := os.Executable()
	if err != nil {
		log.Error("无法获取可执行文件路径。")
		exit.LauncherExit(err)
	}
	dirname := filepath.Dir(ex)
	dirname = filepath.Join(dirname, "launcher_data")
	err = os.MkdirAll(dirname, os.ModePerm)
	if err != nil {
		log.Error("无法创建数据目录。")
		exit.LauncherExit(err)
	}
	_ = os.Remove(filepath.Join(dirname, "account_data.enc"))
	_ = os.Remove(filepath.Join(dirname, "updater.exe"))
	profileFile, err := os.Open(filepath.Join(dirname, "account.enc"))

	if err != nil {
		if os.IsNotExist(err) {
			log.Warn("账户数据文件不存在，开始登录流程。")
			profile = auth.NewLogin()
			SaveAccount(profile)
			return profile
		}
		log.Warn("无法打开账户数据文件。")
		exit.LauncherExit(err)
	}
	profileBytes, err := io.ReadAll(profileFile)
	if err != nil {
		log.Warn("读取账户数据时出现问题，开始登录流程。")
		profile = auth.NewLogin()
		SaveAccount(profile)
		return profile
	}

	_ = profileFile.Close()
	copy(key[:], keyStr)
	nonce := LoadNonce()
	salsa20.XORKeyStream(profileBytes, profileBytes, nonce, &key)
	buf := bytes.NewBuffer(profileBytes)
	decoder := gob.NewDecoder(buf)

	err = decoder.Decode(&profile)
	if err != nil {
		log.Warn("无法解析账户数据，开始登录流程。")
		profile = auth.NewLogin()
		SaveAccount(profile)
	}
	return profile
}
