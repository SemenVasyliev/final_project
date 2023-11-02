package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"time"

	// "github.com/gorilla/sessions"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type Article struct {
	Id                                    int
	Title, Description, ArticleText, Tags string
	Comments                              []Comment
	UserId                                int
	CreatedAt                             string
}

type User struct {
	Id                           int
	Name, Password, Email, Token string
}

type Comment struct {
	ID        int
	PostID    int
	Author    string
	Text      string
	Timestamp time.Time
}

type ArticleWithAuthor struct {
	Article
	AuthorName string
}

// var posts = []Article{}
var showPost = Article{}
var secretKey string = "220203"
var isAuthenticated bool

func CheckAuthentication(tokenString string) bool {

	if tokenString == "" {
		return false
	}

	_, err := checkToken(tokenString)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

func index(w http.ResponseWriter, r *http.Request) {
	// connect files
	cookie, err := r.Cookie("token")
	if err != nil {
		isAuthenticated = false
	} else {
		token := cookie.Value
		isAuthenticated = CheckAuthentication(token)
	}

	t, err := template.ParseFiles("templates/index.html", "templates/header.html", "templates/footer.html", "templates/header_for_auth.html")

	if err != nil {
		fmt.Fprintf(w, err.Error())
	}

	db, err := sql.Open("mysql", "root:220203ctyz@tcp(127.0.0.1:3306)/news")
	if err != nil {
		panic(err)
	}

	defer db.Close()

	data := struct {
		IsAuthenticated bool
		Article         []ArticleWithAuthor
	}{
		IsAuthenticated: isAuthenticated,
	}

	res, err := db.Query(`select a.id, a.title, a.description, a.articleText, a.tags, a.UserId, u.name, a.CreatedAt 
	from users u 
	inner join articles a
	on u.id = a.UserId;`)
	if err != nil {
		panic(err)
	}

	// to make articles empty
	data.Article = []ArticleWithAuthor{}

	for res.Next() {
		var post ArticleWithAuthor
		err = res.Scan(&post.Id, &post.Title, &post.Description, &post.ArticleText, &post.Tags, &post.UserId, &post.AuthorName, &post.CreatedAt)
		if err != nil {
			panic(err)
		}

		data.Article = append(data.Article, post)
	}

	t.ExecuteTemplate(w, "index", data)
}

func create(w http.ResponseWriter, r *http.Request) {
	if isAuthenticated {
		t, err := template.ParseFiles("templates/create.html", "templates/footer.html", "templates/header_for_auth.html")

		if err != nil {
			panic(err)
		}

		t.ExecuteTemplate(w, "create", nil)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

}

func save_article(w http.ResponseWriter, r *http.Request) {
	// taking data from page create.html (<input type="text" name="title" id="title" placeholder="Article title" class="form-control">)
	title := r.FormValue("title")
	description := r.FormValue("description")
	articleText := r.FormValue("articleText")
	tags := r.FormValue("tags")

	cookie, err := r.Cookie("token")
	if err != nil {
		// Обработка ошибки, если токен отсутствует
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tokenString := cookie.Value
	claims, err := checkToken(tokenString)
	if err != nil {
		// Обработка ошибки валидации токена
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userId, _ := claims["user_id"].(float64)
	fmt.Println("User id:")
	fmt.Println(userId)
	// if !ok {
	// 	// Обработка ошибки, если UserId не является int
	// 	fmt.Fprintf(w, "Latif")
	// 	http.Redirect(w, r, "/login", http.StatusSeeOther)
	// 	return
	// }
	intUserId := int(userId)

	//delete
	fmt.Println(intUserId)
	// todo: add a nice check
	if title == "" || description == "" || articleText == "" {
		fmt.Fprintf(w, "Not all data are filled")
	} else {
		db, err := sql.Open("mysql", "root:220203ctyz@tcp(127.0.0.1:3306)/news")
		if err != nil {
			panic(err)
		}

		defer db.Close()
		createdAt := time.Now().Format("2006-01-02 15:04:05")
		// adding to db
		insert, err := db.Query(fmt.Sprintf("INSERT INTO `articles` (`title`, `description`, `articleText`, `tags`, `UserId`, `CreatedAt`) VALUES('%s', '%s', '%s', '%s', '%v', '%s')", title, description, articleText, tags, intUserId, createdAt))
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
		err = res.Scan(&post.Id, &post.Title, &post.Description, &post.ArticleText, &post.Tags, &post.UserId)
		if err != nil {
			panic(err)
		}

		showPost = post
	}

	comments := []Comment{} // Предположим, у вас есть структура Comment для хранения комментариев
	commentsQuery := fmt.Sprintf("SELECT * FROM `comments` WHERE `id` = '%s'", vars["id"])
	commentsResult, err := db.Query(commentsQuery)
	if err != nil {
		panic(err)
	}

	for commentsResult.Next() {
		var comment Comment
		var timestampBytes []byte
		err = commentsResult.Scan(&comment.ID, &comment.PostID, &comment.Author, &comment.Text, &timestampBytes)
		if err != nil {
			panic(err)
		}

		// Преобразование timestampBytes в time.Time
		comment.Timestamp, err = time.Parse("2006-01-02 15:04:05", string(timestampBytes))
		if err != nil {
			panic(err)
		}

		comments = append(comments, comment)
	}

	showPost.Comments = comments

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

		var userId int
		err = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userId)
		if err != nil {
			panic(err)
		}

		// to delete
		fmt.Println(userId)

		cooke := http.Cookie{
			Name:     "token",
			Value:    generateToken(userId),
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		}
		http.SetCookie(w, &cooke)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func generateToken(userId int) string {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userId
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		panic(err)
	}

	return tokenString
}
func checkToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Invalid token signing method")
		}
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("Invalid token")
}

// func GetUserIDFromToken(tokenString string) (int, error) {
// 	claims, err := checkToken(tokenString)
// 	if err != nil {
// 		return -1, err // Возвращаем -1 и ошибку в случае проблемы с токеном
// 	}

// 	// Предполагая, что UserId хранится в токене как строка
// 	userIdStr, ok := claims["UserId"].(string)
// 	if !ok {
// 		return -1, fmt.Errorf("UserId not found in token")
// 	}

// 	userId, err := strconv.Atoi(userIdStr)
// 	if err != nil {
// 		return -1, err
// 	}

// 	return userId, nil
// }

func addComment(w http.ResponseWriter, r *http.Request) {
	if isAuthenticated {
		postID := r.FormValue("post_id")
		commentText := r.FormValue("comment_text")
		author := "CurrentLoggedInUser" // Замените на имя текущего аутентифицированного пользователя

		// Проверка наличия текста комментария
		if commentText == "" {
			http.Error(w, "Comment text is required", http.StatusBadRequest)
			return
		}

		db, err := sql.Open("mysql", "root:220203ctyz@tcp(127.0.0.1:3306)/news")
		if err != nil {
			panic(err)
		}

		defer db.Close()
		// Вставка комментария в базу данных
		_, error := db.Exec("INSERT INTO comments (post_id, author, text, timestamp) VALUES (?, ?, ?, NOW())", postID, author, commentText)
		if error != nil {
			http.Error(w, "Failed to add comment", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/post/%s", postID), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
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
	rtr.HandleFunc("/save_user", save_user).Methods("POST")
	rtr.HandleFunc("/login_user", login_user).Methods("POST")
	rtr.HandleFunc("/add_comment", addComment).Methods("POST")
	rtr.HandleFunc("/logout", logout).Methods("POST")

	http.ListenAndServe(":8081", nil)

}

func logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "token",
		Value:  "",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	handleFunc()
}
