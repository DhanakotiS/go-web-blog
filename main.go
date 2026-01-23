package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	uuid "github.com/google/uuid"

	constants "go-web-blog/constants"
	dbops "go-web-blog/dbops"
	models "go-web-blog/models"
)

var PORT = constants.PORT
var BLOGSFILE = constants.BLOGSFILE

type Blog = models.Blog
type User = models.User

func getMethod(w http.ResponseWriter, r *http.Request, page string) {
	cookie, err := r.Cookie("username")
	if err != nil {
		http.ServeFile(w, r, page)

	} else if cookie.Value != "" {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	} else {
		http.ServeFile(w, r, page)
	}

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func setCookie(w http.ResponseWriter, creds User) {
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

func hashString(str string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(str))
	return h.Sum32()
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/signup" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
	switch r.Method {
	case "GET":
		getMethod(w, r, "static/signup.html")
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		password := r.FormValue("password")
		encodedPassword := base64.StdEncoding.EncodeToString([]byte(password))

		user := User{
			Name:     r.FormValue("name"),
			Username: r.FormValue("username"),
			Email:    r.FormValue("email"),
			Password: encodedPassword,
			UserID:   hashString(r.FormValue("username")),
		}
		insertId, err := dbops.Write("users", &user)
		if err != nil {
			http.Redirect(w, r, "/signup", http.StatusBadRequest)
		}
		log.Printf("User data inserted with ID %s\n", &insertId)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "%d Not Found", http.StatusNotFound)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/login" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		getMethod(w, r, "static/login.html")

	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		creds := User{
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
		}
		setCookie(w, creds)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
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
	blogsData, err := os.ReadFile(BLOGSFILE)
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
		userid := hashString(cookie.Value)
		blog := Blog{
			UserID:  userid,
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

		err = os.WriteFile(BLOGSFILE, jsonData, 0644)
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

func dbHandler(w http.ResponseWriter, r *http.Request) {
	dbops.Init()
}

func main() {

	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file:", err)
	}
	defer file.Close() // Ensure the file is closed when the main function exits

	// Set the standard logger's output to the file
	log.SetOutput(file)

	mux := http.NewServeMux()

	mux.HandleFunc("/", indexHandler)
	// mux.HandleFunc("/signup", signupHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/signup", signupHandler)

	mux.HandleFunc("/home", homeHandler)
	mux.HandleFunc("/create", createHandler)
	mux.HandleFunc("/dbping", dbHandler)
	stop := make(chan os.Signal, 1)
	// listen for interrupt, terminate signal and store in chan
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// dbops.Init()

	go func() {
		fmt.Printf("Starting server at Port %s\n", PORT)

		fs := http.FileServer(http.Dir("static"))
		mux.Handle("/static/", http.StripPrefix("/static/", fs))
		if err := http.ListenAndServe(PORT, mux); err != nil {
			log.Fatal(err)
		}
	}()
	<-stop

	fmt.Println("\nShutting down...")
}
