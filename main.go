package main

import (
    "fmt"
    "log"
    "net/http"
    "container/list"
    "html/template"
    "github.com/bradfitz/gomemcache/memcache"
    "github.com/gorilla/mux"
    "time"
    "encoding/binary"
)

type User struct {
    username string
    password string
    following *list.List
    follower *list.List
    messages *list.List
}

type Message struct {
    timestamp time.Time
    text string
}

func registerHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Println("method:", request.Method)//get request method
    if request.Method == "GET" {
        t, _ := template.ParseFiles("register.gtpl")
        t.Execute(response, nil)
    } else {
        request.ParseForm()
        fmt.Println("username:", request.Form["username"])
        mc := memcache.New("127.0.0.1:11211")
        username := request.FormValue("username")
        password := request.FormValue("password")
        user := &User{}
        user.password = password
        //TODO  Encode struct to byte array

        
        val, err := mc.Get("username")
        if err == nil {
            fmt.Fprintf(response, "<script>alert('this username has been registered')</script>")
            t, _ := template.ParseFiles("register.gtpl")
            t.Execute(response, nil)
        } else {
            fmt.Println(err)
            mc.Set(&memcache.Item{Key: username, Value: []byte(user)})
            t, _ := template.ParseFiles("home.gtpl")
                t.Execute(response, nil)
        }
    }
}

func loginHandler(response http.ResponseWriter, request *http.Request) {
    if request.Method == "GET" {
        t, _ := template.ParseFiles("login.gtpl")
        t.Execute(response, nil)
    } else {
        request.ParseForm()
        fmt.Println("username:", request.Form["username"])
        mc := memcache.New("127.0.0.1:11211")
        username := request.FormValue("username")
        password := request.FormValue("password")

        val, err := mc.Get("username")
        //TODO  Decode byte array to struct  
        if err == nil {
            if password == val.Value {
                t, _ := template.ParseFiles("home.gtpl")
                t.Execute(response, nil)
            } else {
                t, _ := template.ParseFiles("login.gtpl")
                t.Execute(response, nil)
            }
        }


    }
}

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "This is a test for the broadcasting system %s!", r.URL.Path[1:])
}

var router = mux.NewRouter()

func main() {
    mc := memcache.New("127.0.0.1:11211")
    mc.Set(&memcache.Item{Key: "key_one", Value: []byte("michael")})
    mc.Set(&memcache.Item{Key: "key_two", Value: []byte("programming")})
    val, err := mc.Get("key_one")

    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Printf("-- %s", val.Value)
    http.Handle("/", router)
    router.HandleFunc("/", handler)
    router.HandleFunc("/register", registerHandler)
    router.HandleFunc("/login", loginHandler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}