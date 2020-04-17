package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"mtconfig"
	"mtdealer"
	"mtlog"
	"mttraderapi/controller"
	_ "mttraderapi/docs"
	"mttraderapi/httputil"
	"mttraderapi/model"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/sherifabdlnaby/configuro"
	"go.uber.org/zap"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title MT4 Trader REST API
// @version 1.0
// @description This is a mt4 trader rest api server.
// @termsOfService http://swagger.io/terms/

// @contact.name devtraders
// @contact.url https://dev4traders.com
// @contact.email mikhail@dev4traders.com

// @license.name Commerce

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

type Config struct {
	WebServer struct {
		Addr string
		Mode string
	}

	JWT struct {
		Realm      string
		Secret     string
		Timeout    time.Duration
		MaxRefresh time.Duration
	}

	ManagerLogger mtconfig.Common

	Manager mtdealer.Config
}

//var mu sync.Mutex
var dealer *mtdealer.DealerManager
var manager *mtdealer.MarketManager
var conf *Config
var log *zap.Logger

func initLog() error {

	// loading mt4 api manager logger
	l, err := mtlog.NewLogger(conf.ManagerLogger.LogPath, conf.ManagerLogger.LogLevel)
	if err != nil {
		return err
	}

	mtlog.SetDefault(l)
	mtlog.Info("log path: \"%s\" with level \"%s\"", conf.ManagerLogger.LogPath, conf.ManagerLogger.LogLevel)
	// -- loading mt4 api manager logger

	// loading zap logger
	var cfg zap.Config
	f, _ := ioutil.ReadFile(".\\logger.json")
	if err := json.Unmarshal(f, &cfg); err != nil {
		panic(err)
	}
	if log, err = cfg.Build(); err != nil {
		panic(err)
	}
	// -- loading zap logger

	return nil
}
func main() {

	configLoader, err := configuro.NewConfig()
	if err != nil {
		panic(err)
	}

	conf = &Config{}

	if err := configLoader.Load(conf); err != nil {
		panic(err)
	}

	if err := initLog(); err != nil {
		panic(err)
	}

	dealer = mtdealer.NewDealerManager(&conf.Manager)
	dealer.Start()

	manager = mtdealer.NewMarketManager(&conf.Manager)
	manager.Start()

	defer func() {
		_ = log.Sync()
		manager.Stop()
		dealer.Stop()

	}()

	gin.SetMode(conf.WebServer.Mode)
	router := gin.Default()
	c := controller.NewController()

	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       conf.JWT.Realm,
		Key:         []byte(conf.JWT.Secret),
		Timeout:     conf.JWT.Timeout,
		MaxRefresh:  conf.JWT.MaxRefresh,
		IdentityKey: model.KEY_LOGIN,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*model.User); ok {
				return jwt.MapClaims{
					model.KEY_LOGIN: v.Login,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &model.User{
				Login: claims[model.KEY_LOGIN].(int),
			}
		},
		Authenticator: c.UserAuth,
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			httputil.NewError(c, http.StatusUnauthorized, errors.New(message))
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})

	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	v1 := router.Group("/api/v1")
	{
		v1.Use(loadDealerCtx())

		auth := v1.Group("/auth")
		{
			auth.POST("/login", authMiddleware.LoginHandler)
			// Refresh time can be longer than token timeout
			auth.GET("/refresh_token", authMiddleware.RefreshHandler)
		}

		trades := v1.Group("/trades")
		{
			trades.Use(authMiddleware.MiddlewareFunc())
			trades.GET(":login", c.ListUserTrades)
			trades.POST("add", c.AddTrade)
			trades.PATCH("update", c.UpdateTrade)
			trades.PATCH("close", c.CloseTrade)
		}

	}
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if err := router.Run(conf.WebServer.Addr); err != nil {
		panic(err)
	}

}

func loadDealerCtx() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		ctx.Set(model.KEY_MANAGER, manager)
		ctx.Set(model.KEY_DEALER, dealer)

		ctx.Next()
	}
}
