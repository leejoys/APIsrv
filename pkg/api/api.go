package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const NewsOnPage = 15

//Модель полной формы новости
type NewsFullDetailed struct {
	ID      int    `xml:"-" json:"ID"`                // номер записи
	Title   string `xml:"title" json:"Title"`         // заголовок публикации
	Content string `xml:"description" json:"Content"` // содержание публикации
	PubDate string `xml:"pubDate" json:"-"`           // время публикации из RSS
	PubTime int64  `xml:"-" json:"PubTime"`           //время публикации для БД и фронта
	Link    string `xml:"link" json:"Link"`           // ссылка на источник
}

//Модель короткой формы новости
type NewsShortDetailed struct {
	ID      int    `xml:"-" json:"ID"`                // номер записи
	Title   string `xml:"title" json:"Title"`         // заголовок публикации
	Content string `xml:"description" json:"Content"` // содержание публикации
	PubDate string `xml:"pubDate" json:"-"`           // время публикации из RSS
	PubTime int64  `xml:"-" json:"PubTime"`           //время публикации для БД и фронта
	Link    string `xml:"link" json:"Link"`           // ссылка на источник
}

type Paginator struct {
	SumOfPages  int
	CurrentPage int
	NewsOnPage  int
}

type DBAnswer struct {
	Count int
	Posts []NewsShortDetailed
}

type GWAnswer struct {
	PostsArr  []NewsShortDetailed
	Paginator Paginator
}

// Comment - комментарий.
type Comment struct {
	ID            int    `json:"ID"`            // номер записи
	Author        string `json:"Author"`        // автор комментария
	Content       string `json:"Content"`       // содержание комментария
	PubTime       int64  `json:"PubTime"`       //время комментария для БД и фронта
	ParentPost    int    `json:"ParentPost"`    // ID родительской новости
	ParentComment int    `json:"ParentComment"` // ID родительского комментария
}

// Программный интерфейс приложения
type API struct {
	r *mux.Router
}

// Конструктор объекта API
func New() *API {
	api := API{}
	api.r = mux.NewRouter()
	api.endpoints()
	return &api
}

// Регистрация обработчиков API.
func (api *API) endpoints() {

	//метод вывода списка новостей
	api.r.HandleFunc("/news/latest", api.latest).Methods(http.MethodGet)
	//метод фильтра новостей
	api.r.HandleFunc("/news/filter", api.filter).Methods(http.MethodGet)
	//метод получения детальной новости
	api.r.HandleFunc("/news/detailed", api.detailed).Methods(http.MethodGet)
	//метод добавления комментария
	api.r.HandleFunc("/comments/store", api.storeComment).Methods(http.MethodPost)
	// //метод получения комментариев по id новости.
	// api.r.HandleFunc("/comments/{id}", api.comments).Methods(http.MethodGet)
}

// Получение маршрутизатора запросов.
// Требуется для передачи маршрутизатора веб-серверу.
func (api *API) Router() *mux.Router {
	return api.r
}

//метод вывода списка новостей
//localhost:8080/news/latest?page=1
func (api *API) latest(w http.ResponseWriter, r *http.Request) {
	page := 0
	pageS := r.URL.Query().Get("page")
	if pageS != "" {
		var err error
		page, err = strconv.Atoi(pageS)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:8081/news/%d/%d", page, NewsOnPage))
	if err != nil {
		http.Error(w, fmt.Sprintf("latest http.Get error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	bPosts, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("latest ReadAll error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != 200 {
		http.Error(w, string(bPosts), http.StatusInternalServerError)
		return
	}

	dba := DBAnswer{}
	err = json.Unmarshal(bPosts, &dba)
	if err != nil {
		http.Error(w, fmt.Sprintf("latest Unmarshal error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	answer := GWAnswer{}
	answer.PostsArr = dba.Posts
	answer.Paginator.NewsOnPage = NewsOnPage
	answer.Paginator.CurrentPage = page
	answer.Paginator.SumOfPages = int(math.Ceil(float64(dba.Count) / float64(NewsOnPage)))
	bytes, err := json.Marshal(answer)
	if err != nil {
		http.Error(w, fmt.Sprintf("latest Marshal error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

//метод фильтра новостей
//localhost:8080/news/filter?page=1&keyword=Go
func (api *API) filter(w http.ResponseWriter, r *http.Request) {
	page := 0
	q := r.URL.Query()
	k := q.Get("keyword")
	pageS := r.URL.Query().Get("page")
	if pageS != "" {
		var err error
		page, err = strconv.Atoi(pageS)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:8081/filter/%d/%d/%s", page, NewsOnPage, k))
	if err != nil {
		http.Error(w, fmt.Sprintf("filter http.Get error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	bPosts, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("filter ReadAll error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != 200 {
		http.Error(w, string(bPosts), http.StatusInternalServerError)
		return
	}

	dba := DBAnswer{}
	err = json.Unmarshal(bPosts, &dba)
	if err != nil {
		http.Error(w, fmt.Sprintf("filter Unmarshal error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	answer := GWAnswer{}
	answer.PostsArr = dba.Posts
	answer.Paginator.NewsOnPage = NewsOnPage
	answer.Paginator.CurrentPage = page
	answer.Paginator.SumOfPages = int(math.Ceil(float64(dba.Count) / float64(NewsOnPage)))
	bytes, err := json.Marshal(answer)
	if err != nil {
		http.Error(w, fmt.Sprintf("filter Marshal error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

func (api *API) storeComment(w http.ResponseWriter, r *http.Request) {
	bComment, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("APIsrv storeComment ReadAll error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	c := Comment{}
	err = json.Unmarshal(bComment, &c)
	if err != nil {
		http.Error(w, fmt.Sprintf("APIsrv storeComment Unmarshal error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	resp, err := http.Post("http://127.0.0.1:8082/comments", "JSON", bytes.NewReader(bComment))
	if err != nil {
		http.Error(w, fmt.Sprintf("storeComment http.Post error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
}

//метод получения детальной новости,
//localhost:8080/news/detailed?id=1
func (api *API) detailed(w http.ResponseWriter, r *http.Request) {
	id := 0
	idS := r.URL.Query().Get("id")
	if idS != "" {
		var err error
		id, err = strconv.Atoi(idS)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:8081/news/%d/%d", id))
	if err != nil {
		http.Error(w, fmt.Sprintf("latest http.Get error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	bPosts, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("latest ReadAll error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != 200 {
		http.Error(w, string(bPosts), http.StatusInternalServerError)
		return
	}

	dba := DBAnswer{}
	err = json.Unmarshal(bPosts, &dba)
	if err != nil {
		http.Error(w, fmt.Sprintf("latest Unmarshal error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	answer := GWAnswer{}
	answer.PostsArr = dba.Posts
	answer.Paginator.NewsOnPage = NewsOnPage
	answer.Paginator.CurrentPage = page
	answer.Paginator.SumOfPages = int(math.Ceil(float64(dba.Count) / float64(NewsOnPage)))
	bytes, err := json.Marshal(answer)
	if err != nil {
		http.Error(w, fmt.Sprintf("latest Marshal error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

// func (api *API) detailed(w http.ResponseWriter, r *http.Request) {
// 	idS := r.URL.Query().Get("id")
// 	id, err := strconv.Atoi(idS)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	post, err := api.newsDB.PostsDetailedN(id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	bytes, err := json.Marshal(post)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.Write(bytes)
// }

//метод получения комментария по id.
//localhost:8080/comments/1
// func (api *API) comments(w http.ResponseWriter, r *http.Request) {
// 	idS := mux.Vars(r)["id"]
// 	id, err := strconv.Atoi(idS)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}
// 	comments, err := api.commentsDB.Comments(id)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	bytes, err := json.Marshal(comments)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.Write(bytes)
// }
