package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/tkanos/gonfig"
)

type Article struct {
	Id                            uint16
	Title, Announcement, FullText string
}

type Configuration struct {
	Port   int
	paswrd string
	user   string
}

var posts = []Article{}

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html", "templates/top_menu.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	db, err := sql.Open("mysql", "root:******@tcp(127.0.0.1:3306)/test_lucky")
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

func save_article(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	announcement := r.FormValue("announcement")
	full_text := r.FormValue("full_text")

	if title == "" || announcement == "" || full_text == "" {
		fmt.Fprintf(w, "Required parameters not set")
	} else {
		db, err := sql.Open("mysql", "root:******@tcp(127.0.0.1:3306)/test_lucky")
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
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.HandleFunc("/", index)
	http.HandleFunc("/create", create)
	http.HandleFunc("/save_article", save_article)
	http.ListenAndServe(":8080", nil)
}

func main() {
	configuration := Configuration{}
	err := gonfig.GetConf("./cfg/main.cfg", &configuration)
	if err != nil {
		panic(err)
	}
	handleFunc()
}
