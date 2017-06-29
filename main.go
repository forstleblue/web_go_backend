package main

import (
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/unirep/ur-local-web/app/config"
	_ "github.com/unirep/ur-local-web/app/routes"
)

var appConfig = config.Config()
var appCache = cache.New(30*time.Minute, time.Minute)

//var emitter *utils.EventEmitter

//auto run before main
func init() {

	//TODO: create AppInit event here
	//emitter = utils.NewEventEmitter()
	//emitter.Emit("App.OnInit", nil, false)

	// utils.Emitter.On("App.OnInit", func(arg interface{}) interface{} {
	// 	//log.Println("Starting Listener")
	// 	return nil
	// })
}

func main() {

	//initWebsockets()
	//initOauth()
	//routes.AddRoutes()
}
