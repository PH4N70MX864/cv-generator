package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func setupResponse(w *http.ResponseWriter, r *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	if (*r).Method == "OPTIONS" {
		return
	}
	switch r.Method {
	case "POST":
		decoder := json.NewDecoder(r.Body)
		var user User
		err := decoder.Decode(&user)
		if err != nil {
			w.Header().Set("Content-type", "application/json")
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"message": "Bad request"}`))
			return
		}
		fname, err := createFile(user)
		if err != nil {
			w.Header().Set("Content-type", "application/json")
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "Failed creating file"}`))
			return
		}

		// Open file
		f, err := os.Open(fname)
		if err != nil {
			w.Header().Set("Content-type", "application/json")
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "Failed processing file"}`))
			return
		}
		defer f.Close()
		// Set header
		w.Header().Set("Content-type", "application/pdf")

		// Stream to response
		if _, err := io.Copy(w, f); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "Failed sending file"}`))
			return
		}

		e := os.RemoveAll(filepath.Dir(fname))
		if e != nil {
			log.Fatal(e)
		}
	default:
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message": "Method not allowed"}`))
		return
	}

}

func main() {
	http.HandleFunc("/user", serveFile)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
