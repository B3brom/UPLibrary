package main

import (
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"clothingStore/internal/config"
	"clothingStore/internal/db"
	"clothingStore/internal/handler"
	"clothingStore/internal/logger"
	mw "clothingStore/internal/middleware"
	"clothingStore/internal/repository"
	"clothingStore/internal/router"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)

	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		log.Error("database directory", "err", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Error("uploads dir", "err", err)
		os.Exit(1)
	}

	database, err := db.Connect(cfg.DBPath)
	if err != nil {
		log.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Error("migration failed", "err", err)
		os.Exit(1)
	}
	productRepo := repository.NewProductRepo(database)
	photoRepo := repository.NewPhotoRepo(database)
	roleRepo := repository.NewRoleRepo(database)
	userRepo := repository.NewUserRepo(database)
	cartRepo := repository.NewCartRepo(database)
	orderRepo := repository.NewOrderRepo(database)
	orderItemRepo := repository.NewOrderItemRepo(database)

	r := router.New(router.Deps{
		Product:   handler.NewProductHandler(productRepo),
		Photo:     handler.NewPhotoHandler(photoRepo, cfg.UploadDir),
		Role:      handler.NewRoleHandler(roleRepo),
		User:      handler.NewUserHandler(userRepo),
		Cart:      handler.NewCartHandler(cartRepo),
		Order:     handler.NewOrderHandler(orderRepo),
		OrderItem: handler.NewOrderItemHandler(orderItemRepo, orderRepo),
		Auth:      handler.NewAuthHandler(userRepo, cfg.JWTSecret, cfg.JWTTTL),
		JWTSecret: cfg.JWTSecret,
		LogMw:     mw.Logger(log),
		UploadDir: cfg.UploadDir,
	})

	listener, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		log.Error("server start failed", "err", err)
		os.Exit(1)
	}

	url := "http://localhost:" + cfg.Port
	go func() {
		if err := openBrowser(url); err != nil {
			log.Warn("could not open browser", "url", url, "err", err)
		}
	}()

	log.Info("server started", "url", url)
	if err := http.Serve(listener, r); err != nil {
		log.Error("server stopped", "err", err)
	}
}

func openBrowser(url string) error {
	var command string
	var args []string

	switch runtime.GOOS {
	case "windows":
		command = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		command = "open"
		args = []string{url}
	default:
		command = "xdg-open"
		args = []string{url}
	}

	return exec.Command(command, args...).Run()
}
