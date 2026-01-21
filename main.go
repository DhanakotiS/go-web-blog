package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	uuid "github.com/google/uuid"
)

var BlogsFile string = "blogs.json"

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type Blog struct {
	UserID  string `json:"userid"`
	BlogID  string `json:"blogid"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func setCookie(w http.ResponseWriter, creds Credentials) {
	expiration := time.Now().Add(60 * time.Minute)
	userCookie := http.Cookie{
		Name:     "username",
		Value:    creds.Username,
		Expires:  expiration,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	}
	http.SetCookie(w, &userCookie)
}
func unsetCookie(w http.ResponseWriter) {
	userCookie := &http.Cookie{
		Name:     "username",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, userCookie)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		cookie, err := r.Cookie("username")
		if err != nil {
			http.ServeFile(w, r, "static/login.html")

		} else {
			http.Redirect(w, r, "/home", http.StatusSeeOther)
		}
		fmt.Println(cookie)
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		creds := Credentials{
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
		}

		setCookie(w, creds)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		// jsonData, err := json.MarshalIndent(creds, "", " ")
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// err = os.WriteFile("users.json", jsonData, 0644)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
	default:
		fmt.Fprintf(w, "Method %s is not supported.", r.Method)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	unsetCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	type homeData struct {
		User string
	}
	if r.URL.Path != "/home" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	userCookie, err := r.Cookie("username")
	if err != nil {
		// If the cookie is missing or invalid, redirect to login
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	loggedInUser := userCookie.Value

	data := homeData{
		User: loggedInUser,
	}
	tmp, err := template.ParseFiles("static/home.html")
	if err != nil {
		http.Error(w, "Template Parsing error", http.StatusInternalServerError)
		log.Printf("Template parsing error: %v", err)
	}
	err = tmp.Execute(w, data)
	if err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func readDB() ([]Blog, error) {
	// var blogsData []Blog
	blogsData, err := os.ReadFile(BlogsFile)
	// fmt.Printf("%T", blogsData)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var blogs []Blog
	err = json.Unmarshal(blogsData, &blogs)
	if err != nil {
		return nil, err
	}
	return blogs, nil

}

func createHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/create" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	switch r.Method {
	case "GET":
		// Serve from "static" directory for consistency
		http.ServeFile(w, r, "static/blog.html")
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		title := r.FormValue("title")
		content := r.FormValue("content")
		id, err := uuid.NewRandom()
		if err != nil {
			http.Error(w, "Could not generate Blog ID", http.StatusInternalServerError)
			return
		}

		cookie, err := r.Cookie("username")
		user := cookie.Value
		blog := Blog{
			UserID:  user,
			Title:   title,
			Content: content,
			BlogID:  id.String(),
		}

		blogs, err := readDB()
		if err != nil {
			http.Error(w, "Internal read error", http.StatusInternalServerError)
			return
		}

		blogs = append(blogs, blog)

		jsonData, err := json.MarshalIndent(blogs, "", "	")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = os.WriteFile(BlogsFile, jsonData, 0644)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// fmt.Fprintf(w, "%s request success\n", r.Method)
		http.Redirect(w, r, "/home", http.StatusFound)
	default:
		fmt.Fprintf(w, "Method %s is not supported.", r.Method)
	}

}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)

	mux.HandleFunc("/home", homeHandler)
	mux.HandleFunc("/create", createHandler)

	fmt.Printf("Starting server at Port :8080\n")

	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
