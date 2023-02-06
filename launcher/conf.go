package launcher

import (
	"encoding/json"
	"epminecraft-go/exit"
	"io"
	"net/http"
)

var confRemoteUrl = "https://modget.maxng.cc:8086/" + GetUpdateChan() + "/launcher.json"

type Conf struct {
	Latest struct {
		Version int    `json:"version"`
		URL     string `json:"url"`
	} `json:"latest"`
	GameVersion string   `json:"game_version"`
	IsPure      bool     `json:"is_pure"`
	JvmArgs     []string `json:"jvm_args"`
}

func GetConf() Conf {
	client := http.Client{}
	resp, err := client.Get(confRemoteUrl)
	if err != nil {
		log.Error("无法连接到更新服务器。")
		exit.LauncherExit(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			exit.LauncherExit(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		exit.LauncherExit(err)
	}

	var remoteConf Conf
	err = json.Unmarshal(body, &remoteConf)
	if err != nil {
		exit.LauncherExit(err)
	}

	return remoteConf

}
