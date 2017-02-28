package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/tiaanl/go-api/pkg/products"
)

var (
	mySigningKey = []byte("secret")
)

var ProductsHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(products.Products)
})

var AddFeedbackHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	slug := vars["slug"]

	var product products.Product

	for _, p := range products.Products {
		if p.Slug == slug {
			product = p
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if product.Slug != "" {
		json.NewEncoder(w).Encode(product)
	} else {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Product not found",
		})
	}
})

var GetTokenHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["admin"] = true
	claims["name"] = "Some Name"
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	tokenString, _ := token.SignedString(mySigningKey)

	w.Write([]byte(tokenString))
})

var validationKeyGetter = func(token *jwt.Token) (interface{}, error) {
	return mySigningKey, nil
}

var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: validationKeyGetter,
	SigningMethod:       jwt.SigningMethodHS256,
})

func main() {
	router := mux.NewRouter()

	router.Handle("/", http.FileServer(http.Dir("./views/")))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	router.Handle("/api/products", jwtMiddleware.Handler(ProductsHandler)).Methods("GET")
	router.Handle("/api/products/{slug}/feedback", jwtMiddleware.Handler(AddFeedbackHandler)).Methods("GET")

	router.Handle("/auth/token", GetTokenHandler).Methods("GET")

	http.ListenAndServe(":3000", handlers.LoggingHandler(os.Stdout, router))
}
