package main

import (
	"audit-proxy-gateway/internal/app"
)

func main() {
	appInstance := app.InitApp()
	app.StartApp(appInstance)
}
