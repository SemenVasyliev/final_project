package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Article struct {
	Id                                    int
	Title, Description, ArticleText, Tags string
}

var posts = []Article{}
var showPost = Article{}

func index(w http.ResponseWriter, r *http.Request) {
	// connect files
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	db, err := sql.Open("mysql", "root:220203ctyz@tcp(127.0.0.1:3306)/news")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	res, err := db.Query("SELECT * FROM `articles`")
	if err != nil {
		panic(err)
	}

	// to make articles empty
	posts = []Article{}

	for res.Next() {
		var post Article
		err = res.Scan(&post.Id, &post.Title, &post.Description, &post.ArticleText, &post.Tags)
		if err != nil {
			panic(err)
		}

		posts = append(posts, post)
	}

	t.ExecuteTemplate(w, "index", posts)
}

func create(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/create.html", "templates/header.html", "templates/footer.html")

	if err != nil {
		panic(err)
	}

	t.ExecuteTemplate(w, "create", nil)
}

func save_article(w http.ResponseWriter, r *http.Request) {
	// taking data from page create.html (<input type="text" name="title" id="title" placeholder="Article title" class="form-control">)
	title := r.FormValue("title")
	description := r.FormValue("description")
	articleText := r.FormValue("articleText")
	tags := r.FormValue("tags")

	// todo: add a nice check
	if title == "" || description == "" || articleText == "" {
		fmt.Fprintf(w, "Not all data are filled")
	} else {
		db, err := sql.Open("mysql", "root:220203ctyz@tcp(127.0.0.1:3306)/news")
		if err != nil {
			panic(err)
		}

		defer db.Close()

		// adding to db
		insert, err := db.Query(fmt.Sprintf("INSERT INTO `articles` (`title`, `description`, `articleText`, `tags`) VALUES('%s', '%s', '%s', '%s')", title, description, articleText, tags))
		if err != nil {
			panic(err)
		}
		defer insert.Close()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func show_post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	t, err := template.ParseFiles("templates/show.html", "templates/header.html", "templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	db, err := sql.Open("mysql", "root:220203ctyz@tcp(127.0.0.1:3306)/news")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	// выборка данных
	res, err := db.Query(fmt.Sprintf("SELECT * FROM `articles` WHERE `id` = '%s'", vars["id"]))
	if err != nil {
		panic(err)
	}

	// to make articles empty
	showPost = Article{}

	for res.Next() {
		var post Article
		err = res.Scan(&post.Id, &post.Title, &post.Description, &post.ArticleText, &post.Tags)
		if err != nil {
			panic(err)
		}

		showPost = post
	}

	t.ExecuteTemplate(w, "show", showPost)

}

func login(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/login.html", "templates/header.html")
	if err != nil {
		panic(err)
	}
	t.ExecuteTemplate(w, "login", nil)

}

func register(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/register.html", "templates/header.html")
	if err != nil {
		panic(err)
	}
	t.ExecuteTemplate(w, "register", nil)

}

func handleFunc() {
	rtr := mux.NewRouter()
	http.Handle("/", rtr)

	rtr.HandleFunc("/", index).Methods("GET")
	rtr.HandleFunc("/create", create).Methods("GET")
	rtr.HandleFunc("/save_article", save_article).Methods("POST")
	rtr.HandleFunc("/post/{id:[0-9]+}", show_post).Methods("GET") // ,"POST"
	rtr.HandleFunc("/login", login).Methods("GET")
	rtr.HandleFunc("/register", register).Methods("GET")

	http.ListenAndServe(":8081", nil)

}

func main() {
	handleFunc()
}
