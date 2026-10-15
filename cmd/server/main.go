package main

import "github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"

func main() {
	appCfg := config.New()

	appCfg.Logger.Info("hello world")

}
