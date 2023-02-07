package launcher

import (
	"encoding/json"
	"epminecraft-go/exit"
	"fmt"
	"github.com/cespare/xxhash"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/schollz/progressbar/v3"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var modsRemoteUrl = "https://modget.maxng.cc:8086/" + GetUpdateChan() + "/mods.json"

type Mod struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
	URL  string `json:"url,omitempty"`
}

type ModConf struct {
	Required []Mod
	Rejected []Mod
}

// difference returns the elements in `a` that aren't in `b`.
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func getModConf() ModConf {
	client := http.Client{}
	resp, err := client.Get(modsRemoteUrl)
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

	var remoteConf ModConf
	err = json.Unmarshal(body, &remoteConf)
	if err != nil {
		exit.LauncherExit(err)
	}

	return remoteConf
}

func getLocalMods() []Mod {
	ex, err := os.Executable()
	if err != nil {
		exit.LauncherExit(err)
	}
	dirname := filepath.Dir(ex)
	dirname = filepath.Join(dirname, ".minecraft", "mods")
	files, err := os.ReadDir(dirname)
	if err != nil {
		exit.LauncherExit(err)
	}
	var localMods []Mod
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".jar") {
			fileObj, _ := os.Open(filepath.Join(dirname, file.Name()))
			fileBytes, _ := io.ReadAll(fileObj)
			hash := xxhash.Sum64(fileBytes)
			localMods = append(localMods, Mod{Name: file.Name(), Hash: fmt.Sprintf("%016s", strconv.FormatUint(hash, 16))})
		}
	}
	return localMods
}

func downloadMod(mod Mod) {
	client := http.Client{}

	if mod.URL == "" {
		mod.URL = "https://modget.maxng.cc:8086/" + GetUpdateChan() + "/mods/" + mod.Name
	}

	resp, err := client.Get(mod.URL)
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

	ex, err := os.Executable()
	if err != nil {
		exit.LauncherExit(err)
	}
	dirname := filepath.Dir(ex)
	dirname = filepath.Join(dirname, ".minecraft", "mods")
	file, err := os.Create(filepath.Join(dirname, mod.Name))
	if err != nil {
		exit.LauncherExit(err)
	}

	_, _ = io.Copy(file, resp.Body)

	defer func(File *os.File) {
		err := File.Close()
		if err != nil {
			exit.LauncherExit(err)
		}
	}(file)
}

func CheckRequiredMods() {
	remoteConf := getModConf()
	localMods := getLocalMods()
	var lackMods []Mod
	var ReqModHashes []string
	for _, mod := range remoteConf.Required {
		ReqModHashes = append(ReqModHashes, mod.Hash)
	}
	var LocalModHashes []string
	for _, mod := range localMods {
		LocalModHashes = append(LocalModHashes, mod.Hash)
	}
	diffModHashes := difference(ReqModHashes, LocalModHashes)
	for _, mod := range remoteConf.Required {
		for _, hash := range diffModHashes {
			if mod.Hash == hash {
				lackMods = append(lackMods, mod)
			}
		}
	}
	if len(lackMods) > 0 {
		log.Warn("以下Mod未安装或版本不正确：")
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Mod Name", "Mod Hash"})
		for _, mod := range lackMods {
			t.AppendRow(table.Row{mod.Name, mod.Hash})
		}
		t.AppendSeparator()
		t.Render()
		log.Info("正在下载缺失Mod...")
		progress := progressbar.Default(int64(len(lackMods)))
		for _, mod := range lackMods {
			downloadMod(mod)
			progress.Add(1)
		}
		progress.Finish()
	} else {
		log.Info("所有必要Mod已安装。")
	}
}

func CheckRejectedMods() {
	remoteConf := getModConf()
	localMods := getLocalMods()
	var rejectMods []Mod
	var RejModHashes []string
	for _, mod := range remoteConf.Rejected {
		RejModHashes = append(RejModHashes, mod.Hash)
	}
	var localModHashes []string
	for _, mod := range localMods {
		localModHashes = append(localModHashes, mod.Hash)
	}

	for _, mod := range remoteConf.Rejected {
		for _, hash := range localModHashes {
			if mod.Hash == hash {
				rejectMods = append(rejectMods, mod)
			}
		}
	}
	if len(rejectMods) > 0 {
		log.Warn("服务器要求删除以下Mod：")
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Mod Name", "Mod Hash"})
		for _, mod := range rejectMods {
			t.AppendRow(table.Row{mod.Name, mod.Hash})
		}
		t.AppendSeparator()
		t.Render()

		progress := progressbar.Default(int64(len(rejectMods)))

		for _, mod := range rejectMods {
			f := filepath.Join(".minecraft", "mods", mod.Name)
			err := os.Remove(f)
			if err != nil {
				exit.LauncherExit(err)
			}
			_ = progress.Add(1)
		}
		_ = progress.Finish()

	} else {
		log.Info("客户端不存在不允许安装的Mod。")
	}
}
