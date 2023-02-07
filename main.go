package main

import (
	"epminecraft-go/auth"
	"epminecraft-go/exit"
	"epminecraft-go/launcher"
	"epminecraft-go/logger"
	"epminecraft-go/storage"
	"epminecraft-go/updater"
	"fmt"
	"os"
)

var log = logger.Logger()

func main() {

	log.Info(fmt.Sprintf("MaxLauncher 版本:%d", launcher.GetSelfVer()) + " 初始化中...")
	log.Info(fmt.Sprintf("更新通道: %s", launcher.GetUpdateChan()))
	remoteConf := launcher.GetConf()
	if launcher.GetSelfVer() < remoteConf.Latest.Version {
		log.Info("发现新版本启动器！版本: " + fmt.Sprintf("%d", remoteConf.Latest.Version))
		log.Info("正在更新启动器...")
		updater.DoUpdate(remoteConf.Latest.URL)
		log.Info("启动器更新完成，将在下次启动时生效。")
	} else {
		log.Info("你正在使用最新版本的启动器。")
	}
	mcAccount := storage.LoadAccount()
	mcAccount = auth.RefreshLogin(mcAccount)
	storage.SaveAccount(mcAccount)
	launchArgs := launcher.ComposeArgs(mcAccount)

	log.Info("欢迎，" + mcAccount.Name + "!")

	if !remoteConf.IsPure {
		log.Info("配置为模组服，开始模组列表检查...")
		launcher.CheckRejectedMods()
		launcher.CheckRequiredMods()
		log.Info("模组列表检查完成。")
	} else {
		log.Info("配置为纯净服，跳过模组列表检查。")
	}

	log.Info("正在启动 Minecraft...")

	procAttr := new(os.ProcAttr)
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	if process, err := os.StartProcess(launchArgs[0], launchArgs[1:], procAttr); err != nil {
		fmt.Printf("ERROR Unable to run Minecraft: %s\n", err.Error())
	} else {
		fmt.Printf("Minecraft running as pid %d\n", process.Pid)
		_, err = process.Wait()
		if err != nil {
			log.Fatalf("wait: %v", err)
		}
	}
	exit.LauncherExit(nil)
}
