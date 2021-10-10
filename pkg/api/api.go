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

//Модель новости
type News struct {
	ID      int    `xml:"-" json:"ID"`                // номер записи
	Title   string `xml:"title" json:"Title"`         // заголовок публикации
	Content string `xml:"description" json:"Content"` // содержание публикации
	PubDate string `xml:"pubDate" json:"-"`           // время публикации из RSS
	PubTime int64  `xml:"-" json:"PubTime"`           //время публикации для БД и фронта
	Link    string `xml:"link" json:"Link"`           // ссылка на источник
}

// объект пагинации
type Paginator struct {
	SumOfPages  int
	CurrentPage int
	NewsOnPage  int
}

//объект ответа БД новостей
type DBAnswer struct {
	Count int
	Posts []News
}

//Модель короткой формы новостей
type NewsShortDetailed struct {
	PostsArr  []News
	Paginator Paginator
}

//Модель полной формы новости
type NewsFullDetailed struct {
	Post        News
	CommentsArr []Comment
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
	answer := NewsShortDetailed{}
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
	answer := NewsShortDetailed{}
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

	cens, err := http.Post("http://127.0.0.1:8083/cens", "text", bytes.NewReader([]byte(c.Content)))
	if err != nil {
		http.Error(w, fmt.Sprintf("storeComment censPost error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	bCens, err := io.ReadAll(cens.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("filter ReadAll error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if cens.StatusCode != 200 {
		http.Error(w, string(bCens), http.StatusInternalServerError)
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

	type newsDTO struct {
		news News
		err  error
	}
	type commentsDTO struct {
		comments []Comment
		err      error
	}

	detailedFunc := func(ch chan newsDTO) {
		nDTO := newsDTO{}
		det, err := http.Get(fmt.Sprintf("http://127.0.0.1:8081/detailed/%d", id))
		if err != nil {
			nDTO.news = News{}
			nDTO.err = fmt.Errorf("detailed http.Get 8081 error: %s", err)
			ch <- nDTO
			return
		}

		bPost, err := io.ReadAll(det.Body)
		if err != nil {
			nDTO.news = News{}
			nDTO.err = fmt.Errorf("detailed ReadAll 8081 error: %s", err)
			ch <- nDTO
			return
		}

		if det.StatusCode != 200 {
			nDTO.news = News{}
			nDTO.err = fmt.Errorf("detailed !=200 8081 error: %s", string(bPost))
			ch <- nDTO
			return
		}

		p := News{}
		err = json.Unmarshal(bPost, &p)
		if err != nil {
			nDTO.news = News{}
			nDTO.err = fmt.Errorf("detailed Unmarshal 8081 error: %s", err)
			ch <- nDTO
			return
		}
		nDTO.news = p
		nDTO.err = nil
		ch <- nDTO
	}
	commentsFunc := func(ch chan commentsDTO) {
		cDTO := commentsDTO{}
		com, err := http.Get(fmt.Sprintf("http://127.0.0.1:8082/comments/%d", id))
		if err != nil {
			cDTO.comments = []Comment{}
			cDTO.err = fmt.Errorf("detailed http.Get 8082 error: %s", err)
			ch <- cDTO
			return
		}

		bComms, err := io.ReadAll(com.Body)
		if err != nil {
			cDTO.comments = []Comment{}
			cDTO.err = fmt.Errorf("detailed ReadAll 8082 error: %s", err)
			ch <- cDTO
			return
		}

		if com.StatusCode != 200 {
			cDTO.comments = []Comment{}
			cDTO.err = fmt.Errorf("detailed !=200 8082 error: %s", string(bComms))
			ch <- cDTO
			return
		}

		c := []Comment{}
		err = json.Unmarshal(bComms, &c)
		if err != nil {
			cDTO.comments = []Comment{}
			cDTO.err = fmt.Errorf("detailed Unmarshal 8082 error: %s", err)
			ch <- cDTO
			return
		}
		cDTO.comments = c
		cDTO.err = nil
		ch <- cDTO
	}

	answer := NewsFullDetailed{}
	newschan := make(chan newsDTO, 1)
	commentschan := make(chan commentsDTO, 1)
	go detailedFunc(newschan)
	go commentsFunc(commentschan)

	for i := 0; i < 2; i++ {
		select {
		case ndto := <-newschan:
			{
				if ndto.err != nil {
					http.Error(w, fmt.Sprintf("detailed newschan error: %s", ndto.err.Error()), http.StatusInternalServerError)
					return
				}
				answer.Post = ndto.news
			}
		case cdto := <-commentschan:
			{
				if cdto.err != nil {
					http.Error(w, fmt.Sprintf("detailed commentschan error: %s", cdto.err.Error()), http.StatusInternalServerError)
					return
				}
				answer.CommentsArr = cdto.comments
			}
		}
	}
	bytes, err := json.Marshal(answer)
	if err != nil {
		http.Error(w, fmt.Sprintf("detailed Marshal error: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	w.Write(bytes)
}

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
