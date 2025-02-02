package supaauth

import (
	"context"
	"github.com/buger/jsonparser"
	"github.com/go-logr/logr"
	"github.com/robinbraemer/event"
	"github.com/spf13/viper"
	"github.com/supabase-community/supabase-go"
	"go.minekube.com/common/minecraft/color"
	"go.minekube.com/common/minecraft/component"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/util/uuid"
	"strconv"
	"strings"
	"sync"
	"time"
)

var client supabase.Client
var log logr.Logger

type User struct {
	Id       string
	Ip       string
	Username string
	Password string
}

var users map[string]User
var usersMutex sync.Mutex
var usersLastUpdated time.Time
var usersLastUpdateMutex sync.Mutex

func getUserByIP(ip string) *User {
	for _, u := range users {
		if u.Ip == ip {
			return &u
		}
	}
	return nil
}

func getUserByUsername(username string) *User {
	user := users[username]
	if user.Username == "" {
		return nil
	}
	return &user
}

func getUserByPassword(password string) *User {
	for _, u := range users {
		if u.Password == password {
			return &u
		}
	}
	return nil
}

var Plugin = proxy.Plugin{
	Name: "supaauth",
	Init: func(ctx context.Context, p *proxy.Proxy) error {
		log = logr.FromContextOrDiscard(ctx)
		configInit()
		if !viper.GetBool("enable") {
			return nil
		}

		c, err := supabase.NewClient(
			viper.GetString("supabaseApiUrl"),
			viper.GetString("supabaseApiKey"),
			&supabase.ClientOptions{},
		)
		if err != nil {
			return err
		}

		client = *c
		go fetchUsers()

		event.Subscribe(p.Event(), 0, onLogin)
		event.Subscribe(p.Event(), 0, onPing)
		event.Subscribe(p.Event(), 0, onGameProfileRequest)

		log.Info("supaauth plugin enabled")
		return nil
	},
}

func fetchUsers() {
	for {
		data, _, err := client.From("minecraft").Select("*", "exact", false).Execute()
		if err != nil {
			log.Error(err, "Unable to fetchUsers")
			return
		}

		localUsers := make(map[string]User)

		_, _ = jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			id, _ := jsonparser.GetString(value, "id")
			ip, _ := jsonparser.GetString(value, "ip")
			username, _ := jsonparser.GetString(value, "username")
			password, _ := jsonparser.GetString(value, "password")

			localUsers[username] = User{id, ip, username, password}
		})
		usersMutex.Lock()
		usersLastUpdateMutex.Lock()

		users = localUsers
		usersLastUpdated = time.Now()

		usersMutex.Unlock()
		usersLastUpdateMutex.Unlock()

		time.Sleep(1 * time.Minute)
	}
}

func configInit() {
	viper.SetConfigName("supaauth")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")

	viper.SetDefault("enable", false)
	viper.SetDefault("enableIpWhitelist", false)
	viper.SetDefault("supabaseApiKey", "")
	viper.SetDefault("supabaseApiUrl", "")

	if err := viper.ReadInConfig(); err != nil {
		log.Info("Error reading config file, using defaults")
	}
}

func onGameProfileRequest(e *proxy.GameProfileRequestEvent) {
	password := strings.Split(e.Conn().VirtualHost().String(), ".")[0]
	user := getUserByPassword(password) // user verified

	profile := e.GameProfile()

	if viper.GetBool("changePlayerUUID") {
		parsedUUID, _ := uuid.Parse(user.Id)
		profile.ID = parsedUUID
	}

	if viper.GetBool("changePlayerUsername") {
		profile.Name = user.Username
	}
	e.SetGameProfile(profile)
}

func onLogin(e *proxy.LoginEvent) {
	if viper.GetBool("enableIpWhitelist") {
		// an IP filter because there may be users with one IP address
		user := getUserByIP(strings.Split(e.Player().RemoteAddr().String(), ":")[0])
		if user == nil {
			e.Deny(&component.Text{
				Content: "Вас нет в вайтлисте",
				S:       component.Style{Color: color.DarkRed},
			})
			return
		}
	}
	password := strings.Split(e.Player().VirtualHost().String(), ".")[0]
	user := getUserByPassword(password)
	if user == nil {
		toUpdate := int((time.Now().Sub(usersLastUpdated)).Seconds())
		e.Deny(&component.Text{
			Content: "Обновите адрес входа на сайте auranite.ru!\n" +
				"Или попробуйте подождать " + strconv.Itoa(60-toUpdate) + " с.",
		})
		return
	}
}

func onPing(e *proxy.PingEvent) {
	original := e.Ping()
	if viper.GetBool("enableIpWhitelist") {
		user := getUserByIP(strings.Split(e.Connection().RemoteAddr().String(), ":")[0])
		if user == nil {
			original.Description = &component.Text{
				Content: "Вас нет в вайтлисте",
			}
			e.SetPing(original)
			return
		}
	}
	password := strings.Split(e.Connection().VirtualHost().String(), ".")[0]
	user := getUserByPassword(password)
	if user == nil {
		toUpdate := int((time.Now().Sub(usersLastUpdated)).Seconds())
		original.Description = &component.Text{
			Content: "Обновите адрес входа на сайте auranite.ru!\n" +
				"Или попробуйте подождать " + strconv.Itoa(60-toUpdate) + " с.",
		}
		e.SetPing(original)
		return
	}
}
