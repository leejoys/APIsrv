/*Реализовать скелет сервиса API Gateway.
Создайте новый проект в IDE, назовите его APIGateway.
Нужно в рамках этого проекта поднять HTTP-сервер для адреса:

http://localhost:8080/

И добавить следующие обработчики:

метод вывода списка новостей,
метод фильтра новостей,
метод получения детальной новости,
метод добавления комментария.

Пока нам не нужно делать полную реализацию. Вам требуется:
*добавить модели NewsFullDetailed, NewsShortDetailed, Comment
*в методах, которые отдают данные, в теле определить объект или массив
(в зависимости от метода) и возвращать эти данные в качестве ответа

То есть вам нужно использовать то, что называется hard-code.
Методы будут возвращать всегда одно и то же, но пока это не важно.
*/

package main

import (
	"apisrv/pkg/api"
	"log"
	"net/http"
)

type server struct {
	api *api.API
}

func main() {
	srv := server{}

	srv.api = api.New()

	// Запускаем веб-сервер на порту 8081 на всех интерфейсах.
	// Предаём серверу маршрутизатор запросов.
	log.Println("HTTP server is started on localhost:8080")
	defer log.Println("HTTP server has been stopped")
	log.Fatal(http.ListenAndServe("localhost:8080", srv.api.Router()))
}
