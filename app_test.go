package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

var a App

func TestMain(m *testing.M){
	_ = godotenv.Load()
	var DBUser = os.Getenv("DBUSER")
	var DBPassword = os.Getenv("DBPASSWORD")
	var DBName = os.Getenv("TESTDBNAME")
	if DBName == "" {
		DBName = "test"
	}
	var Host = os.Getenv("DBHOST")
	err := a.Initialise(DBUser, DBPassword, DBName, Host)
	if err != nil {
		panic(err)
	}
	createTable()
	m.Run()
}

func createTable() {
	createTableQuery := `CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    quantity INT NOT NULL,
    price DOUBLE PRECISION NOT NULL
);`

	_, err := a.DB.Exec(createTableQuery)
	if err != nil {
		panic(err)
	}
}

func clearTable() {
	_, err := a.DB.Exec("DELETE FROM products")
	if err != nil {
		panic(err)
	}
	_, err = a.DB.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1")
	if err != nil {
		panic(err)
	}
}

func addProduct(name string, quantity int, price float64) {
	_, err := a.DB.Exec("INSERT INTO products (name, quantity, price) VALUES ($1, $2, $3)", name, quantity, price)
	if err != nil {
		panic(err)
	}
}

func TestGetProduct(t *testing.T) {
	clearTable()
	addProduct("Test Product", 10, 99.99)
	req, _ := http.NewRequest("GET", "/products/1", nil)
	response := sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)
}

func TestCreateProduct(t *testing.T) {
	clearTable()
	product := []byte(`{"name":"Test Product","quantity":10,"price":99.99}`)
	req, _ := http.NewRequest("POST", "/products", bytes.NewBuffer(product))
	req.Header.Set("Content-Type", "application/json")
	response := sendRequest(req)
	checkStatusCode(t, http.StatusCreated, response.Code)

	var m map[string]interface{}
	json.Unmarshal(response.Body.Bytes(), &m)
	if m["name"] != "Test Product" {
		t.Errorf("Expected product name to be 'Test Product', got '%s'", m["name"])
	}
	if m["quantity"] != 10.0 {
		t.Errorf("Expected product quantity to be 10, got %f", m["quantity"])
	}
	if m["price"] != 99.99 {
		t.Errorf("Expected product price to be 99.99, got %f", m["price"])
	}
}

func TestDeleteProduct(t *testing.T) {
	clearTable()
	addProduct("Test Product", 10, 99.99)
	req, _ := http.NewRequest("DELETE", "/products/1", nil)
	response := sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/products/1", nil)
	response = sendRequest(req)
	checkStatusCode(t, http.StatusNotFound, response.Code)
}

func TestUpdateProduct(t *testing.T) {
	clearTable()
	addProduct("Test Product", 10, 99.99)
	updatedProduct := []byte(`{"name":"Updated Product","quantity":20,"price":199.99}`)
	req, _ := http.NewRequest("PUT", "/products/1", bytes.NewBuffer(updatedProduct))
	req.Header.Set("Content-Type", "application/json")
	response := sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)

	var m map[string]interface{}
	if json.Unmarshal(response.Body.Bytes(), &m) != nil {
		t.Errorf("Error unmarshalling response body")
	}
	if m["name"] != "Updated Product" || m["quantity"] != 20.0 || m["price"] != 199.99 {
		t.Errorf("Expected product to be updated, got %v", m)
	}

	req, _ = http.NewRequest("GET", "/products/1", nil)
	response = sendRequest(req)
	checkStatusCode(t, http.StatusOK, response.Code)

	var p map[string]interface{}
	if json.Unmarshal(response.Body.Bytes(), &p) != nil {
		t.Errorf("Error unmarshalling response body")
	}
	if p["name"] != "Updated Product" || p["quantity"] != 20.0 || p["price"] != 199.99 {
		t.Errorf("Expected product to be updated, got %v", p)
	}
}

func checkStatusCode(t *testing.T, expected int, actual int) {
	if expected != actual {
		t.Errorf("Expected status code %d, got %d", expected, actual)
	}
}

func sendRequest(req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	a.Router.ServeHTTP(recorder, req)
	return recorder
}