package main

import (
	"apisrv/pkg/api"
	"log"
	"net/http"
	"os"
	"os/signal"
)

type server struct {
	api *api.API
}

func main() {
	srv := server{}

	srv.api = api.New()

	// Запускаем веб-сервер на порту 8080 на всех интерфейсах.
	// Предаём серверу маршрутизатор запросов.
	go func() {
		log.Fatal(http.ListenAndServe("localhost:8080", srv.api.Router()))
	}()
	log.Println("HTTP server is started on localhost:8080")
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	<-signalCh
	log.Println("HTTP server has been stopped")
}
