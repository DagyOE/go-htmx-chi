package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/DagyOE/go-htmx-chi/database"
	"github.com/DagyOE/go-htmx-chi/middlewares"
	"github.com/DagyOE/go-htmx-chi/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func init() {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error in loading .env file.")
	}

	database.ConnectDB()
}

func main() {
	defer database.DBConn.Close()
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", homeHandler)
	r.Get("/user-info", userInfoHandler)

	r.Get("/posts", postHandler)

	r.Get("/post/create", createPostHandler)
	r.Post("/post/create", createPostHandler)

	r.Route("/post/{id}", func(r chi.Router) {
		r.Use(middlewares.PostCtx)

		r.Get("/", getPostHandler)

		r.Get("/edit", editPostHandler)
		r.Post("/edit", editPostHandler)

		r.Delete("/delete", deletePostHandler)
	})

	http.ListenAndServe(":3000", r)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {

	ctx := make(map[string]string)

	ctx["name"] = "John Doe"

	t, _ := template.ParseFiles("templates/index.html")

	err := t.Execute(w, ctx)

	if err != nil {
		fmt.Printf("Error in template execution: %v", err)
	}
}

func userInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("User Info from API server"))
}

func createPostHandler(w http.ResponseWriter, r *http.Request) {

	ctx := make(map[string]interface{})

	// Post part

	if r.Method == http.MethodPost {

		r.ParseForm()

		title := r.FormValue("title")
		description := r.FormValue("description")

		stmt := "insert into posts (title, description) values ($1, $2)"

		q, err := database.DBConn.Prepare(stmt)

		if err != nil {
			log.Fatalf("Error inserting data: %q", err)
		}

		res, err := q.Exec(title, description)

		if err != nil {
			log.Fatalf("Error inserting data: %q", err)
		}

		rowsAffected, _ := res.RowsAffected()

		if rowsAffected == 1 {
			ctx["success"] = "Post created successfully."
		}
	}

	// Get part

	t, _ := template.ParseFiles("templates/pages/post_form.html")

	err := t.Execute(w, ctx)

	if err != nil {
		fmt.Printf("Error in template execution: %v", err)
	}
}

func postHandler(w http.ResponseWriter, r *http.Request) {

	var posts []models.Post

	sql := "select * from posts"

	rows, err := database.DBConn.Query(sql)

	if err != nil {
		log.Fatalf("Error querying database: %q", err)
	}

	for rows.Next() {
		data := models.Post{}

		err := rows.Scan(&data.Id, &data.Title, &data.Description)

		if err != nil {
			log.Fatalf("Error scanning rows: %q", err)
		}

		posts = append(posts, data)
	}

	ctx := make(map[string]interface{})

	ctx["posts"] = posts
	ctx["heading"] = "Article List"

	t, _ := template.ParseFiles("templates/pages/post.html")

	err = t.Execute(w, ctx)

	if err != nil {
		fmt.Printf("Error in template execution: %v", err)
	}
}

func editPostHandler(w http.ResponseWriter, r *http.Request) {

	ctx := make(map[string]interface{})
	post := r.Context().Value("post").(models.Post)

	// Post part

	if r.Method == http.MethodPost {

		r.ParseForm()

		title := r.FormValue("title")
		description := r.FormValue("description")

		stmt := "update posts set title=$1, description=$2 where id=$3"

		query, err := database.DBConn.Prepare(stmt)

		if err != nil {
			log.Fatalf("Error in preparing query: %q", err)
		}

		res, err := query.Exec(title, description, post.Id)

		if err != nil {
			log.Fatalf("Error in executing query: %q", err)
		}

		rowsAffected, _ := res.RowsAffected()

		if rowsAffected == 1 {
			ctx["success"] = "Post updated successfully."
		}

	}

	// Load template

	t, _ := template.ParseFiles("templates/pages/post_form.html")

	ctx["post"] = post
	err := t.Execute(w, ctx)

	if err != nil {
		log.Println("Error in tpl execution", err)
	}
}

func deletePostHandler(w http.ResponseWriter, r *http.Request) {

	post := r.Context().Value("post").(models.Post)

	stmt := "delete from posts where id=$1"

	query, err := database.DBConn.Prepare(stmt)

	catchErr(err)

	res, err := query.Exec(post.Id)

	catchErr(err)

	rowsAffected, _ := res.RowsAffected()

	if rowsAffected == 1 {
		http.Redirect(w, r, "/posts", http.StatusSeeOther)
	}
}

func getPostHandler(w http.ResponseWriter, r *http.Request) {

	post := r.Context().Value("post")

	t, _ := template.ParseFiles("templates/pages/post_detail.html")

	ctx := make(map[string]interface{})
	ctx["post"] = post
	err := t.Execute(w, ctx)

	if err != nil {
		log.Println("Error in tpl execution", err)
	}

}

func catchErr(err error) {
	if err != nil {
		log.Fatalf("Error: %q", err)
	}
}
