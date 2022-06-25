package main

//"fmt"
import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"gitlab.com/mgdi/kongroo-c2/c2/config"
	mongo "gitlab.com/mgdi/kongroo-c2/c2/database/mongo"
	redis "gitlab.com/mgdi/kongroo-c2/c2/database/redis"

	"gitlab.com/mgdi/kongroo-c2/c2/router"
)

var (
	err error
)

func init() {
	log.SetLevel(log.DebugLevel)

	config.InitializeConfigs()

	// Initialize mongodb connection using NewClient method
	err = mongo.NewClient(fmt.Sprintf("mongodb://%s:%s", config.Configs["mongo.host"], config.Configs["mongo.port"]), config.Configs["mongo.database"])
	if err != nil {
		log.Fatal(err)
	}
	if err := mongo.MongoCl.CreateAllAgentsCollection(); err != nil {
		log.Fatal(err)
	}
	redis.NewClient()
}

// @title Kongroo-C2 APIs docs
// @version 1.0
// @description kongroo-c2 documentation for APIs
// @BasePath /
// @contact.name 0x1337
// @contact.email info@kongroo.c2
// @host localhost
func main() {
	go router.Run()

	sc := make(chan os.Signal, 1)

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	appCleanup()
	os.Exit(1)
}

func appCleanup() {
	log.Println("Gracefully shutdown...")
	mongo.MongoCl.CloseClient()

}
