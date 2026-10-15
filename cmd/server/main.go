package main

import (
	"fmt"

	"github.com/agar-ayamsegar/agar-data-collection/internal/db/mongo"
	"github.com/agar-ayamsegar/agar-data-collection/internal/utils/config"
)

func main() {
	appCfg := config.New()
	db := mongo.New(appCfg)

	fmt.Println(db.DSN())
}
