package updater

import (
	"epminecraft-go/exit"
	"epminecraft-go/logger"
	"github.com/minio/selfupdate"
	"io"
	"net/http"
	"time"
)

var log = logger.Logger()

func DoUpdate(url string) {
	resp, err := http.Get(url)
	if err != nil {
		exit.LauncherExit(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	err = selfupdate.Apply(resp.Body, selfupdate.Options{})

	if err != nil {
		exit.LauncherExit(err)
	}

	log.Info("更新完成，请再次打开启动器以应用更新。")
	time.Sleep(2 * time.Second)
	log.Info("继续启动流程...")

}
