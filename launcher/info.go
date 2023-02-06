package launcher

import (
	"epminecraft-go/exit"
	"strconv"
)

var selfVer string
var updateChan string

func GetSelfVer() int {
	res, err := strconv.Atoi(selfVer)
	if err != nil {
		exit.LauncherExit(err)
	}
	return res
}

func GetUpdateChan() string {
	return updateChan
}
