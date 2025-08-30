package cmd

import (
	"github.com/imedcl/manager-api/pkg/config"
	routes "github.com/imedcl/manager-api/pkg/routes"
)

func Start() {
	cfg := config.New()
	routes.Create(cfg)
}
