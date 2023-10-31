package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func index(w http.ResponseWriter, r *http.Request) {
	// connect files
	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	t.ExecuteTemplate(w, "index", nil)
}

func create(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/create.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		fmt.Fprintf(w, err.Error())
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

func handleFunc() {
	http.HandleFunc("/", index)
	http.HandleFunc("/create", create)
	http.HandleFunc("/save_article", save_article)
	http.ListenAndServe(":8081", nil)
}

func main() {
	handleFunc()
}
