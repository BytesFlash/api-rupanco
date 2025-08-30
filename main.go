package main

import (
	api "github.com/imedcl/manager-api/cmd"
)

// @title Autentia Manager Documentation
// @version 1
// @Description api autentia manager
// @securityDefinitions.apikey BarerToken
// @in header
// @name Authorization

func main() {
	api.Start(nil)
}
