package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/tkanos/gonfig"
)

type Article struct {
	Id                            uint16
	Title, Announcement, FullText string
}

type Configuration struct {
	DB_USERNAME string
	DB_PASSWORD string
	DB_PORT     string
	DB_HOST     string
	DB_NAME     string
}

var connectionString string
var posts = []Article{}
var showPost = Article{}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html", "templates/top_menu.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	defer db.Close()

	result, err := db.Query("select * from `articles`")
	if err != nil {
		panic(err)
	}

	posts = []Article{}
	for result.Next() {
		var post Article
		err = result.Scan(&post.Id, &post.Title, &post.Announcement, &post.FullText)
		if err != nil {
			panic(err)
		}

		posts = append(posts, post)
		// fmt.Println(fmt.Sprintf("Post: %s with id: %d", post.Title, post.Id))
	}

	t.ExecuteTemplate(w, "index", posts)
}

func create(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/create.html", "templates/header.html", "templates/footer.html", "templates/top_menu.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "create", nil)
}

func show_post(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/create.html", "templates/header.html", "templates/footer.html", "templates/top_menu.html", "templates/show.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
	}
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	// SELECT articles --------------------------------------------------------
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		panic(err)
	}

	defer db.Close()

	selectQuery := fmt.Sprintf("SELECT * FROM `articles` WHERE id = '%s'", vars["id"])
	result, err := db.Query(selectQuery)
	if err != nil {
		panic(err)
	}

	showPost = Article{}
	for result.Next() {
		var post Article
		err = result.Scan(&post.Id, &post.Title, &post.Announcement, &post.FullText)
		if err != nil {
			panic(err)
		}

		showPost = post
	}

	t.ExecuteTemplate(w, "show", showPost)
}

func save_article(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	announcement := r.FormValue("announcement")
	full_text := r.FormValue("full_text")

	if title == "" || announcement == "" || full_text == "" {
		fmt.Fprintf(w, "Required parameters not set")
	} else {

		db, err := sql.Open("mysql", connectionString)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			// panic(err)
		}

		defer db.Close()

		insert, err := db.Query(fmt.Sprintf("insert into `articles` (`title`,`announcement`,`full_text`) values ('%s','%s','%s')", title, announcement, full_text))
		if err != nil {
			fmt.Fprintf(w, err.Error())
			// panic(err)
		}

		defer insert.Close()
	}

	// fmt.Fprintf(w,"('%s','%s','%s')",title,announcement,full_text)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleFunc() {

	rtr := mux.NewRouter()
	rtr.HandleFunc("/", index).Methods("GET")
	rtr.HandleFunc("/create", create).Methods("GET")
	rtr.HandleFunc("/save_article", save_article).Methods("POST")
	rtr.HandleFunc("/post/{id:[0-9]+}", show_post).Methods("GET")

	http.Handle("/", rtr)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.ListenAndServe(":8080", nil)

}

func main() {
	configuration := Configuration{}
	err := gonfig.GetConf("./cfg/db.cfg", &configuration)
	if err != nil {
		panic(err)
	}

	connectionString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", configuration.DB_USERNAME, configuration.DB_PASSWORD, configuration.DB_HOST, configuration.DB_PORT, configuration.DB_NAME)
	// fmt.Fprintf(w, "%s", connectionString)
	handleFunc()
}
