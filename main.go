package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"test/handlers"
	"test/models"
	"test/utility"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/astaxie/beego/orm"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

var key []byte
var store = sessions.NewCookieStore([]byte("secret-key"))

func init() {
	err := godotenv.Load()
	if err != nil {
		return
	}
	key = []byte(os.Getenv("ENC_KEY"))
	port := os.Getenv("DB_PORT")
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	conn_str := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, port, dbName)

	fmt.Println("conn_str:", conn_str)
	err = orm.RegisterDataBase("default", "mysql", conn_str)
	if err != nil {
		fmt.Println("there is some error:", err)
	}
	fmt.Println("db connected")
}

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/api/auth/signup", signup).Methods("POST")
	router.HandleFunc("/api/auth/login", login).Methods("POST")
	router.HandleFunc("/api/notes", authenticate(rateLimit(handlers.GetNotes))).Methods("GET")
	router.HandleFunc("/api/notes/{id}", authenticate(rateLimit(handlers.GetNote))).Methods("GET")
	router.HandleFunc("/api/notes", authenticate(rateLimit(handlers.CreateNote))).Methods("POST")
	router.HandleFunc("/api/notes/{id}", authenticate(rateLimit(handlers.UpdateNote))).Methods("PUT")
	router.HandleFunc("/api/notes/{id}", authenticate(rateLimit(handlers.DeleteNoteByID))).Methods("DELETE")
	router.HandleFunc("/api/notes/{id}/share", authenticate(rateLimit(handlers.ShareNoteWithUser))).Methods("POST")
	router.HandleFunc("/api/search", authenticate(rateLimit(handlers.SearchNotes))).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

type visitor struct {
	lastSeen time.Time
	requests int
}

var visitors = make(map[string]*visitor)
var mtx sync.Mutex

func rateLimit(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mtx.Lock()
		defer mtx.Unlock()

		ip := r.RemoteAddr
		v, exists := visitors[ip]
		if !exists {
			visitors[ip] = &visitor{time.Now(), 1}
		} else if time.Since(v.lastSeen) > 1*time.Minute {
			v.lastSeen = time.Now()
			v.requests = 1
		} else {
			v.requests++
			if v.requests > 100 {
				http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func signup(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utility.Respond(http.StatusBadRequest, "wrong format", &w, false)
		return
	}
	var params struct {
		UserName string `json:"username"`
		PassWord string `json:"password"`
	}

	err = json.Unmarshal(body, &params)
	if err != nil {
		utility.Respond(http.StatusBadRequest, "wrong format", &w, false)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(params.PassWord), 8)

	if err != nil {
		http.Error(w, "Unable to parse request body", http.StatusBadRequest)
		return
	}
	_, err = models.GetUserByUserName(params.UserName)
	if err != sql.ErrNoRows {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}

	if err != nil {
		fmt.Println("error:", err)
		return
	}

	user := models.User{Username: params.UserName, Password: string(hashedPassword)}
	err = user.InsertToDB()
	if err != nil {
		http.Error(w, "Unable to parse request body", http.StatusBadRequest)
		return
	}
	JwtHandle(w, r, user)

	w.Write([]byte("User created successfully"))
}

var jwtKey = []byte("your_secret_key")

func login(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utility.Respond(http.StatusBadRequest, "wrong format", &w, false)
		return
	}
	var params struct {
		UserName string `json:"username"`
		PassWord string `json:"password"`
	}

	err = json.Unmarshal(body, &params)
	if err != nil {
		utility.Respond(http.StatusBadRequest, "wrong format", &w, false)
		return
	}

	fmt.Println("params:", params)

	user, err := models.GetUserByUserName(params.UserName)

	if err != nil {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(params.PassWord))
	if err != nil {
		http.Error(w, "Wrong password", http.StatusBadRequest)
		return
	}

	JwtHandle(w, r, user)
	w.Write([]byte("Logged in successfully"))

}

func JwtHandle(w http.ResponseWriter, r *http.Request, user models.User) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &jwt.StandardClaims{
		Subject:   strconv.Itoa(user.Id),
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		http.Error(w, "Error in generating token", http.StatusInternalServerError)
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
		Path:    "/api",
	})
	http.SetCookie(w, &http.Cookie{
		Name:    "UserName",
		Value:   user.Username,
		Expires: expirationTime,
		Path:    "/api",
	})
}

func authenticate(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "Not authenticated", http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tknStr := c.Value
		claims := &jwt.StandardClaims{}

		tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if !tkn.Valid {
			http.Error(w, "Not authenticated", http.StatusUnauthorized)
			return
		}
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				http.Error(w, "Not authenticated", http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}
