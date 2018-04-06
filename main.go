package main

import (
    "fmt"
    "log"
    "net/http"
    //"container/list"
    "html/template"
    "github.com/bradfitz/gomemcache/memcache"
    "github.com/gorilla/mux"
    "time"
    "encoding/gob"
    "bytes"
)

type User struct { // capitalize the first letter of each field name as "exported" for gob 
    Username string
    Password string
    Following map[string]struct{}
    Follower map[string]struct{}
    Messages map[time.Time]string
}

type Message struct {
    timestamp time.Time
    text string
}

func registerHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Println("method:", request.Method) // get request method
    if request.Method == "GET" {
        t, _ := template.ParseFiles("register.gtpl")
        t.Execute(response, nil)
    } else {
        request.ParseForm()
        fmt.Println("username:", request.Form["username"])
        mc := memcache.New("127.0.0.1:11211")
        username := request.FormValue("username")
        password := request.FormValue("password")
        following := make(map[string]struct{})
        follower := make(map[string]struct{})
        user := User {
            Username: username,
            Password: password,
            Following: following,
            Follower: follower,
        }
        
        //TODO  Encode struct to byte array
        encBuf := new(bytes.Buffer)
        encErr := gob.NewEncoder(encBuf).Encode(user)
        if encErr != nil {
            log.Fatal(encErr)
        }
        value := encBuf.Bytes()

     
        _, err := mc.Get(username)
        
        if err == nil {
            fmt.Fprintf(response, "<script>alert('this username has been registered')</script>")
            t, _ := template.ParseFiles("register.gtpl")
            t.Execute(response, nil)
        } else {
            fmt.Println(err)
            mc.Set(&memcache.Item{Key: username, Value: []byte(value)})
            
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

        val, err := mc.Get(username)
        //TODO  Decode byte array to struct
        decBuf := bytes.NewBuffer(val.Value)
        userOut := User{}
        err = gob.NewDecoder(decBuf).Decode(&userOut)

        if err == nil {
            if password == userOut.Password {
                t, _ := template.ParseFiles("home.gtpl")
                t.Execute(response, nil)
            } else {
                fmt.Fprintf(response, "<script>alert('wrong username or password')</script>")
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