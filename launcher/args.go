package launcher

import (
	"epminecraft-go/auth"
	"epminecraft-go/exit"
	"epminecraft-go/logger"
	"errors"
	"fmt"
	"github.com/pbnjay/memory"
	"github.com/tidwall/gjson"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var log = logger.Logger()

func ComposeArgs(profile auth.Profile) []string {
	remoteConf := GetConf()
	gameVer := remoteConf.GameVersion

	ex, err := os.Executable()
	if err != nil {
		exit.LauncherExit(err)
	}
	dirname := filepath.Dir(ex)

	_ = os.Rename(filepath.Join(dirname, "GraalVM_EE"), filepath.Join(dirname, "jre"))
	jvmDirname := filepath.Join(dirname, "jre", "bin", "java.exe")

	var launchArgs []string
	launchArgs = append(launchArgs, jvmDirname)
	launchArgs = append(launchArgs, "-XX:HeapDumpPath=MojangTricksIntelDriversForPerformance_javaw.exe_minecraft.exe.heapdump")

	jvmArgs := remoteConf.JvmArgs
	for _, arg := range jvmArgs {
		launchArgs = append(launchArgs, arg)
	}

	idleMem := float64(memory.FreeMemory() / 1024 / 1024)
	idleMem = idleMem * 0.8
	if idleMem > 8192 {
		idleMem = 8192
	}
	useMem := int(idleMem)
	log.Info(fmt.Sprintf("将使用内存: %d MB", useMem))
	launchArgs = append(launchArgs, fmt.Sprintf("-Xms%dM", useMem))
	launchArgs = append(launchArgs, fmt.Sprintf("-Xmx%dM", useMem))

	dosName, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		log.Warn("无法读取系统信息，跳过此步。")
		launchArgs = append(launchArgs, "-Dos.name=Windows 10")
	} else {
		osName := runtime.GOOS
		majorVer, _, _ := dosName.GetIntegerValue("CurrentMajorVersionNumber")
		minorVer, _, _ := dosName.GetIntegerValue("CurrentMinorVersionNumber")
		launchArgs = append(launchArgs, "-Dos.name="+cases.Title(language.English, cases.Compact).String(osName)+" "+strconv.Itoa(int(majorVer)))
		launchArgs = append(launchArgs, "-Dos.version="+fmt.Sprintf("%d.%d", majorVer, minorVer))
	}
	defer func(dosName registry.Key) {
		err := dosName.Close()
		if err != nil {
			exit.LauncherExit(err)
		}
	}(dosName)

	var libraryPath string
	files, err := os.ReadDir(filepath.Join(".minecraft", "versions", gameVer))
	if err != nil {
		exit.LauncherExit(err)
	}
	for _, file := range files {
		if file.IsDir() && strings.Contains(file.Name(), "natives") {
			libraryPath = file.Name()
			break
		}
	}

	if libraryPath == "" {
		libraryPath = filepath.Join(".minecraft", "libraries")
	}

	launchArgs = append(launchArgs, "-Djava.library.path="+filepath.Join(dirname, ".minecraft", "versions", gameVer, libraryPath))

	launchArgs = append(launchArgs, "-Dminecraft.launcher.brand=MaxLauncher")
	launchArgs = append(launchArgs, "-Dminecraft.launcher.version="+selfVer)
	launchArgs = append(launchArgs, "-cp")

	var allLibs []string
	err = filepath.WalkDir(filepath.Join(dirname, ".minecraft", "libraries"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".jar") {
			allLibs = append(allLibs, path)
		}
		return nil
	})
	if err != nil {
		exit.LauncherExit(err)
	}

	allLibs = append(allLibs, filepath.Join(dirname, ".minecraft", "versions", gameVer, gameVer+".jar"))
	launchArgs = append(launchArgs, strings.Join(allLibs, ";"))

	versionConf, err := os.OpenFile(filepath.Join(dirname, ".minecraft", "versions", gameVer, gameVer+".json"), os.O_RDWR, 0755)
	if err != nil {
		exit.LauncherExit(err)
	}
	defer func(versionConf *os.File) {
		err := versionConf.Close()
		if err != nil {
			exit.LauncherExit(err)
		}
	}(versionConf)

	versionConfBytes, err := io.ReadAll(versionConf)
	if err != nil {
		exit.LauncherExit(err)
	}
	versionConfStr := string(versionConfBytes)

	if err != nil {
		exit.LauncherExit(err)
	}

	if gjson.Get(versionConfStr, "mainClass").String() == "" {
		exit.LauncherExit(errors.New("无法读取游戏版本配置文件"))
	}

	mcArgs := gjson.Get(versionConfStr, "minecraftArguments")
	mcArgsJson := ""
	if mcArgs.Exists() {
		mcArgsJson = mcArgs.String()
	} else {
		gjson.Get(versionConfStr, "arguments.game").ForEach(func(key, value gjson.Result) bool {
			if value.IsObject() {
				return false
			}
			mcArgsJson += value.String() + " "
			return true
		})
	}

	mcArgsJson = strings.ReplaceAll(mcArgsJson, "${version_type}", "release")
	mcArgsJson = strings.ReplaceAll(mcArgsJson, "${auth_player_name}", profile.Name)
	mcArgsJson = strings.ReplaceAll(mcArgsJson, "${version_name}", gameVer)
	mcArgsJson = strings.ReplaceAll(mcArgsJson, "${game_directory}", filepath.Join(dirname, ".minecraft"))
	mcArgsJson = strings.ReplaceAll(mcArgsJson, "${assets_root}", filepath.Join(dirname, ".minecraft", "assets"))
	mcArgsJson = strings.ReplaceAll(mcArgsJson, "${assets_index_name}", gjson.Get(versionConfStr, "assets").String())
	mcArgsJson = strings.ReplaceAll(mcArgsJson, "${auth_uuid}", profile.ID)
	mcArgsJson = strings.ReplaceAll(mcArgsJson, "${auth_access_token}", profile.AccessToken)
	mcArgsJson = strings.ReplaceAll(mcArgsJson, "${user_type}", "mojang")

	if strings.Contains(mcArgsJson, "--tweakClass optifine.OptiFineForgeTweaker ") {
		mcArgsJson = strings.ReplaceAll(mcArgsJson, "--tweakClass optifine.OptiFineForgeTweaker ", "")
		mcArgsJson = mcArgsJson + " --tweakClass optifine.OptiFineForgeTweaker"
	}

	launchArgs = append(launchArgs, gjson.Get(versionConfStr, "mainClass").String())
	launchArgs = append(launchArgs, strings.Split(mcArgsJson, " ")...)

	return launchArgs
}
