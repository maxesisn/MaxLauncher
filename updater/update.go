package updater

import (
	"epminecraft-go/exit"
	"github.com/minio/selfupdate"
	"io"
	"net/http"
)

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
}
