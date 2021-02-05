package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type user struct {
	Email          string `json:"email"`
	HashedPassword string `json:"hashed_password"`
	AccessToken    string `json:"access_token"`
}

// users はuser情報をstoreする。サーバーを再起動するとuser情報はロストする
var users = make([]user, 0)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	b, _ := httputil.DumpRequest(r, true)
	log.Println(string(b))

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := io.WriteString(w, `{"error":"/users/create supports post method only"}"`)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.WriteString(w, `{"error":"failed to read request body"}"`)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	defer r.Body.Close()
	type userJSON struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var param userJSON
	if err := json.Unmarshal(body, &param); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.WriteString(w, `{"error":"failed to unmarshal json"}"`)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	if param.Email == "" || param.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, `{"error":"email and password was required"}"`)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	for _, v := range users {
		if v.Email == param.Email {
			w.WriteHeader(http.StatusBadRequest)
			_, err := io.WriteString(w, `{"error":"user email already exists"}"`)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(param.Password), 10)

	id := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.WriteString(w, `{"error":"failed to issue access token"}"`)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	accessToken := base64.StdEncoding.EncodeToString(id.Bytes())

	u := user{
		Email:          param.Email,
		HashedPassword: string(hashed),
		AccessToken:    accessToken,
	}
	users = append(users, u)

	jsonb, err := json.Marshal(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.WriteString(w, `{"error":"failed to marshal json"}"`)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = io.Copy(w, bytes.NewReader(jsonb))
	if err != nil {
		log.Fatal(err)
	}
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	b, _ := httputil.DumpRequest(r, true)
	log.Println(string(b))

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, err := io.WriteString(w, `{"error":"/users/login supports post method only"}"`)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.WriteString(w, `{"error":"failed to read request body"}"`)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	defer r.Body.Close()
	type userJSON struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var param userJSON
	if err := json.Unmarshal(body, &param); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.WriteString(w, `{"error":"failed to unmarshal json"}"`)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	if param.Email == "" || param.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, `{"error":"email and password was required"}"`)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	var u user
	for _, v := range users {
		if v.Email == param.Email {
			u = v
		}
	}
	// userがいない
	empty := user{}
	if u == empty {
		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, `{"error":"not found user"}"`)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(param.Password))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, err := io.WriteString(w, `{"error":"wrong email or password"}"`)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	jsonb, err := json.Marshal(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.WriteString(w, `{"error":"failed to marshal json"}"`)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, bytes.NewReader(jsonb))
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/users/create", CreateUser)
	http.HandleFunc("/users/login", LoginUser)

	log.Fatal(http.ListenAndServe(":9999", nil))
}
