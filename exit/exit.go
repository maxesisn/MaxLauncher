package exit

import (
	"fmt"
	"github.com/gosuri/uilive"
	"time"
)

func LauncherExit(err error) {
	writer := uilive.New()
	writer.Start()

	stayTime := 3

	if err != nil {
		stayTime = 10
	}

	for i := stayTime; i > 0; i-- {
		if err != nil {
			_, _ = fmt.Fprintf(writer, "启动器遇到问题，错误：%s，将在%d秒后退出。\n", err, i)
		} else {
			_, _ = fmt.Fprintf(writer, "游戏已退出，将在%d秒后退出启动器。\n", i)
		}
		time.Sleep(time.Second)
	}
}
