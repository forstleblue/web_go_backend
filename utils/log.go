package utils

import (
	"log"

	"github.com/fatih/color"
	"github.com/valyala/fasthttp"
)

//Log prints a message in the console
func Log(ctx *fasthttp.RequestCtx, action string, status int, length int, message string) {

	var statColor func(...interface{}) string
	if status > 399 {
		statColor = color.New(color.FgRed, color.BgBlack).SprintFunc()
	} else {
		statColor = color.New(color.FgGreen, color.BgBlack).SprintFunc()
	}

	log.Printf("%v %v %v %v %v %v", statColor(action), string(ctx.Method()), string(ctx.Path()), statColor(status), length, message)
}
