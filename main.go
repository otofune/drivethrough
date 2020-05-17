package main

import (
	"io"
	"log"
	"net/http"

	"github.com/otofune/drivethrough/drive"
)

func main() {
	oauth, err := getOAuth2Config()
	if err != nil {
		log.Fatalf("Unable to setup config: %v", err)
	}
	client, err := getClient(oauth)
	if err != nil {
		log.Fatalf("Unable to setup client: %v", err)
	}

	picker, err := drive.NewFilePicker(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	var h http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		log.Printf("<- %s %s\n", r.Method, r.RequestURI)

		w.Header().Set("via", "drivethrough/0")
		if r.Method != http.MethodGet {
			w.WriteHeader(400)
			io.WriteString(w, "GET only supported.")
			return
		}
		id, err := picker.Lookup(r.URL.Path)
		if err != nil {
			log.Println(err)
			w.WriteHeader(404)
			return
		}
		reader, err := picker.Read(id)
		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}
		defer reader.Close()
		w.WriteHeader(200)
		if _, err := io.Copy(w, reader); err != nil {
			log.Println(err)
			return
		}
	}

	s := &http.Server{
		Addr:    ":10000",
		Handler: h,
	}
	log.Printf("Listening on %s\n", ":10000")
	log.Fatal(s.ListenAndServe())

	return
}
