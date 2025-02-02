package supaauth

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/spf13/viper"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
)

var Plugin = proxy.Plugin{
	Name: "supaauth",
	Init: func(ctx context.Context, p *proxy.Proxy) error {
		log := logr.FromContextOrDiscard(ctx)
		configInit(log)
		log.Info(viper.GetString("enabled"))
		//event.Subscribe(p.Event(), 0, onPreLogin)
		return nil
	},
}

func configInit(log logr.Logger) {
	viper.SetConfigName("supaauth")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")

	viper.SetDefault("enabled", false)
	viper.SetDefault("enableIpWhitelist", false)
	viper.SetDefault("supabaseApi", "")

	if err := viper.ReadInConfig(); err != nil {
		log.Info("Error reading config file, using defaults")
	}
}

func onPreLogin(e *proxy.PreLoginEvent) {
	e.Deny(&component.Text{
		Content: "А вам запрещено",
		S:       component.Style{},
	})
}
