package internal

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nathanjisaac/actual-server-go/internal/core"
	"github.com/nathanjisaac/actual-server-go/internal/routes"
	"github.com/nathanjisaac/actual-server-go/internal/storage"
)

// Used for `SharedArrayBuffer` to work in client
func setHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		c.Response().Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		return next(c)
	}
}

func StartServer(config core.Config, buildDirectory embed.FS, headless bool, logs bool) {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.CORS())
	e.Use(setHeaders)

	if logs {
		e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
			Format: "method=${method}, uri=${uri}, status=${status}\n",
		}))
	}

	if !headless {
		e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
			Root:       "node_modules/@actual-app/web/build",
			HTML5:      true,
			Filesystem: http.FS(buildDirectory),
		}))
	}

	conn, pStore, tStore, fStore, err := storage.NewAccountStores(config.Storage, config.StorageConfig)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer conn.Close()

	handler := routes.RouteHandler{
		Config:        config,
		FileStore:     fStore,
		TokenStore:    tStore,
		PasswordStore: pStore,
	}
	e.GET("/mode", handler.GetMode)

	account := e.Group("/account")
	account.GET("/needs-bootstrap", handler.NeedsBootstrap)
	account.POST("/bootstrap", handler.Bootstrap)
	account.POST("/login", handler.Login)
	account.POST("/change-password", handler.ChangePassword)
	account.GET("/validate", handler.ValidateUser)

	sync := e.Group("/sync")
	sync.POST("/sync", handler.SyncFile)
	sync.POST("/user-create-key", handler.UserCreateKey)
	sync.POST("/user-get-key", handler.UserGetKey)
	sync.POST("/reset-user-file", handler.ResetUserFile)
	sync.POST("/update-user-filename", handler.UpdateUserFileName)
	sync.GET("/get-user-file-info", handler.UserFileInfo)
	sync.GET("/list-user-files", handler.ListUserFiles)
	sync.POST("/upload-user-file", handler.UploadUserFile)
	sync.GET("/download-user-file", handler.DownloadUserFile)
	sync.POST("/delete-user-file", handler.DeleteUserFile)

	e.Logger.Fatal(e.Start(fmt.Sprintf("%v:%v", config.Hostname, config.Port)))
}
