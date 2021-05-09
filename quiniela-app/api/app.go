package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/godror/godror"
	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB     *sql.DB
}

func respondWithMessage(w http.ResponseWriter, code int, message string) {
	response, _ := json.Marshal(map[string]string{"message": message})

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	w.Write(response)
}

func (a *App) Start() {
	databaseDSN, ok := os.LookupEnv("DATABASE_DSN")
	if !ok {
		panic(fmt.Errorf("error opening db: credentials not found"))
	}

	var err error
	a.DB, err = sql.Open("godror", databaseDSN)
	if err != nil {
		panic(fmt.Errorf("error opening db: %w", err))
	}

	a.Router = mux.NewRouter()
	a.initializeRoutes()
}

func (a *App) Run(PORT string) {
	fmt.Println("Servidor corriendo en http://localhost:" + PORT)
	log.Fatal(http.ListenAndServe(":"+PORT, a.Router))
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	u := usuario{Username: username}
	err := u.getUser(a.DB)

	switch err {
	case nil:
		respondWithJSON(w, http.StatusOK, u)
	case sql.ErrNoRows:
		respondWithMessage(w, http.StatusNotFound, "Usuario no existente")
	default:
		respondWithMessage(w, http.StatusInternalServerError, err.Error())
	}
}

//func (a *App) getAllUsers(w http.ResponseWriter, r *http.Request) {
//	limit, _ := strconv.Atoi(r.FormValue("limit"))
//	offset, _ := strconv.Atoi(r.FormValue("offset"))

//	if limit < 1 {
//		limit = 10
//	}
//	if offset < 1 {
//		offset = 0
//	}

//	usuarios, err := getAllUsers(a.DB, offset, limit)
//	if err != nil {
//		respondWithJSON(w, http.StatusInternalServerError, err.Error())
//		return
//	}

//	respondWithList(w, http.StatusOK, usuarios)
//}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	var u usuario
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithMessage(w, http.StatusBadRequest, "Solicitud no válida")
		return
	}
	defer r.Body.Close()

	if err := u.createUser(a.DB); err != nil && err != sql.ErrNoRows {
		respondWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithMessage(w, http.StatusCreated, "Usuario registrado")
}

func (a *App) loginUser(w http.ResponseWriter, r *http.Request) {
	var u usuario
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithMessage(w, http.StatusBadRequest, "Solicitud no válida")
		return
	}
	defer r.Body.Close()

	err := u.loginUser(a.DB)

	switch err {
	case nil:
		respondWithMessage(w, http.StatusOK, "Credenciales válidas")
	case sql.ErrNoRows:
		respondWithMessage(w, http.StatusUnauthorized, "No se encontró esta combinación de usuario y contraseña")
	default:
		respondWithMessage(w, http.StatusInternalServerError, err.Error())
	}
}

func (a *App) recoverUser(w http.ResponseWriter, r *http.Request) {
	var u usuario
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithMessage(w, http.StatusBadRequest, "Solicitud no válida")
		return
	}
	defer r.Body.Close()

	err := u.recoverUser(a.DB)
	switch err {
	case nil:
		respondWithMessage(w, http.StatusOK, u.Contraseña)
	case sql.ErrNoRows:
		respondWithMessage(w, http.StatusUnauthorized, "No se encontró este correo")
	default:
		respondWithMessage(w, http.StatusInternalServerError, err.Error())
	}
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {

	var u usuario
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithMessage(w, http.StatusBadRequest, "Solicitud no válida")
		return
	}
	defer r.Body.Close()

	if err := u.updateUser(a.DB); err != nil && err != sql.ErrNoRows {
		respondWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithMessage(w, http.StatusCreated, "Usuario actualizado ")
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]

	u := usuario{Username: username}
	if err := u.deleteUser(a.DB); err != nil && err != sql.ErrNoRows {
		respondWithMessage(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithMessage(w, http.StatusOK, "Usuario eliminado")
}

func (a *App) initializeRoutes() {
	//a.Router.HandleFunc("/users", a.getAllUsers).Methods("GET")
	a.Router.HandleFunc("/api/user/{username}", corsHandler(a.getUser))
	a.Router.HandleFunc("/api/registrar", corsHandler(a.createUser))
	a.Router.HandleFunc("/api/login", corsHandler(a.loginUser))
	a.Router.HandleFunc("/api/recuperar", corsHandler(a.recoverUser))
	//a.Router.HandleFunc("/user/", corsHandler(a.deleteUser))
}

func corsHandler(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Vary", "Access-Control-Request-Method")
			w.Header().Set("Vary", "Access-Control-Request-Headers")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST,OPTIONS")
		} else {
			handler(w, r)
		}
	}
}