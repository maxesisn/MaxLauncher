package auth

import (
	"epminecraft-go/exit"
	"epminecraft-go/logger"
	"github.com/xmdhs/gomclauncher/auth"
)

var log = logger.Logger()

type Profile = *auth.Profile
type MsToken = *auth.MsToken

func NewLogin() Profile {
	profile, err := auth.MsLogin()
	if err != nil {
		exit.LauncherExit(err)
	}
	return profile
}

func RefreshLogin(profile Profile) Profile {
	profile, err := auth.GetProfile(profile.AccessToken)
	if err != nil {
		log.Warn("登录Token已过期，正在自动刷新...")
		profile, err := auth.MsLoginRefresh(&profile.MsToken)
		if err != nil {
			exit.LauncherExit(err)
		}
		return profile
	}

	if err != nil {
		exit.LauncherExit(err)
	}
	log.Info("登录Token有效，无需刷新。")
	return profile
}
