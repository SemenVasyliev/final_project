package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	// "github.com/gorilla/sessions"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Article struct {
	Id                                    int
	Title, Description, ArticleText, Tags string
}

type User struct {
	Id                    int
	Name, Password, Email string
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

func save_user(w http.ResponseWriter, r *http.Request) {

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if name == "" || email == "" || password == "" {
		fmt.Fprintf(w, "Not all data are filled")

	} else {
		db, err := sql.Open("mysql", "root:220203ctyz@tcp(127.0.0.1:3306)/news")
		if err != nil {
			panic(err)
		}

		defer db.Close()

		// check email
		var existingEmail int
		if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&existingEmail); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if existingEmail > 0 {
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		// adding to db
		insert, err := db.Query(fmt.Sprintf("INSERT INTO `users` (`name`, `password`, `email`) VALUES('%s', '%s', '%s')", name, password, email))
		if err != nil {
			panic(err)
		}
		defer insert.Close()

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func login_user(w http.ResponseWriter, r *http.Request) {

	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		fmt.Fprintf(w, "Not all data are filled")

	} else {
		db, err := sql.Open("mysql", "root:220203ctyz@tcp(127.0.0.1:3306)/news")
		if err != nil {
			panic(err)
		}

		defer db.Close()

		// check email
		var existingEmail int
		var existingPass int
		if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&existingEmail); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if existingEmail > 0 {
			if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE password = ?", password).Scan(&existingPass); err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				return
			}
		}

		if existingPass < 1 {
			//http.Redirect(w, r, "/login", http.StatusSeeOther)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// func sessionCreate(w http.ResponseWriter, r *http.Request, name string) {
// 	session, err := store.Get(r, "session-name") // Создаем или получаем сессию
// if err != nil {
//     http.Error(w, "Session error", http.StatusInternalServerError)
//     return
// }

// // Сохраняем информацию в сессии, например, имя пользователя
// session.Values["user"] = name // Переменную userName необходимо получить после аутентификации

// // Сохраняем сессию
// err = session.Save(r, w)
// if err != nil {
//     http.Error(w, "Session error", http.StatusInternalServerError)
//     return
// }

// // Перенаправляем пользователя на страницу с расширенной функциональностью
// http.Redirect(w, r, "/my_articles", http.StatusSeeOther)
// }

func handleFunc() {
	rtr := mux.NewRouter()
	http.Handle("/", rtr)

	rtr.HandleFunc("/", index).Methods("GET")
	rtr.HandleFunc("/create", create).Methods("GET")
	rtr.HandleFunc("/save_article", save_article).Methods("POST")
	rtr.HandleFunc("/post/{id:[0-9]+}", show_post).Methods("GET") // ,"POST"
	rtr.HandleFunc("/login", login).Methods("GET")
	rtr.HandleFunc("/register", register).Methods("GET")
	rtr.HandleFunc("/save_user", save_user).Methods("POST")
	rtr.HandleFunc("/login_user", login_user).Methods("POST")

	http.ListenAndServe(":8081", nil)

}

func main() {
	handleFunc()
}
