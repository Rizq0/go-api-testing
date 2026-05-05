package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	_ "github.com/jackc/pgx/v5/stdlib"

	"encoding/json"

	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
	DB    *sql.DB
}


func (app *App) Initialise(DBUser string, DBPassword string, DBName string, LocalHost string) error {
	db, err := sql.Open("pgx", "postgres://"+DBUser+":"+DBPassword+"@"+LocalHost+"/"+DBName)
	if err != nil {
		return err
	}
	app.DB = db

	err = app.DB.Ping()
	if err != nil {
		return err
	}

	app.Router = mux.NewRouter().StrictSlash(true)
	app.handleRoutes()
	return nil
}

func (app *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, app.Router))
}

func (app *App) Close() {
	if app.DB != nil {
		app.DB.Close()
	}
}

func sendResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	response, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(response)
}

func sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	sendResponse(w, statusCode, map[string]string{"error": message})
}

func (app *App) getProducts(w http.ResponseWriter, r *http.Request) {
	products, err := getProductsFromDB(app.DB)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve products from database.")
		return
	}
	sendResponse(w, http.StatusOK, products)
}

func (app *App) getProductByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid product ID.")
		return
	}

	p := &Product{ID: id}
	if err := p.getProductByIDFromDB(app.DB); err != nil {
		if err == sql.ErrNoRows {
			sendErrorResponse(w, http.StatusNotFound, "Product not found.")
		} else {
			sendErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve product from database.")
		}
		return
	}
	sendResponse(w, http.StatusOK, p)
}

func (app *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request payload.")
		return
	}
	if err := p.createProductInDB(app.DB); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to create product in database.")
		return
	}
	sendResponse(w, http.StatusCreated, p)
}

func (app *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid product ID.")
		return
	}

	var p Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request payload.")
		return
	}
	p.ID = id
	if err := p.updateProductInDB(app.DB); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to update product in database.")
		return
	}
	sendResponse(w, http.StatusOK, p)
}

func (app *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid product ID.")
		return
	}

	p := Product{ID: id}
	if err := p.deleteProductFromDB(app.DB); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to delete product from database.")
		return
	}
	sendResponse(w, http.StatusOK, map[string]string{"message": "Product deleted successfully."})
}

func (app *App) handleRoutes() {
	app.Router.HandleFunc("/products", app.getProducts).Methods("GET")
	app.Router.HandleFunc("/products/{id}", app.getProductByID).Methods("GET")
	app.Router.HandleFunc("/products", app.createProduct).Methods("POST")
	app.Router.HandleFunc("/products/{id}", app.updateProduct).Methods("PUT")
	app.Router.HandleFunc("/products/{id}", app.deleteProduct).Methods("DELETE")

}

	