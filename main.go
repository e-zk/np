package main

import (
	"database/sql"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	// TODO move these to a config file
	Bullet      = "&bullet;"
	SuperSecret = "super secret key"
	ListenOn    = ":8081"
	TxtFile     = "./nowplaying.db"

	repo *SQLiteRepo
)

func handlePost(w http.ResponseWriter, r *http.Request) {
	var (
		artist = r.PostFormValue("artist")
		track  = r.PostFormValue("track")
		album  = r.PostFormValue("album")
		url    = r.PostFormValue("url")
		secret = r.PostFormValue("key")
	)
	var title string
	var kind int64

	if secret != SuperSecret {
		http.Error(w, "you are forbidden.", http.StatusBadRequest)
		return
	}

	if album != "" {
		title = album
		kind = 1
	} else {
		title = track
		kind = 0
	}

	if artist == "" || title == "" {
		http.Error(w, "artist and either track or album must be specified.", http.StatusBadRequest)
		return
	}

	e := Entry{
		Artist:    artist,
		Title:     title,
		Type:      kind,
		Link:      url,
		DateAdded: time.Now().Format(time.RFC822),
	}
	_, err := repo.Add(e)
	if err != nil {
		log.Fatalf("error adding: %v\n%v\n", e, err)
	}

	return
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	plaintext := false

	// serve plaintext if requested
	if r.Header.Get("Content-Type") == "text/plain" || strings.HasPrefix(r.UserAgent(), "curl") {
		plaintext = true
	}

	// set content-type
	if plaintext {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	a, err := repo.All()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	if !plaintext {
		fmt.Fprintf(w, "<pre>\n")
	}

	for _, i := range a {
		t, err := time.Parse(time.RFC822, i.DateAdded)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(w, "%s %s\t%s - %s\n", Bullet, t.Format("2006-01-02"), i.Artist, i.Title)
	}

	if !plaintext {
		fmt.Fprintf(w, "<pre>\n")
	}

	return
}

func handleRss(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")

	fmt.Fprintf(w, "<rss version=\"2.0\">\n")
	fmt.Fprintf(w, "<channel>\n")
	fmt.Fprintf(w, "<title>music log</title>\n")
	fmt.Fprintf(w, "<description/>\n")

	all, err := repo.All()
	if err != nil {
		log.Fatalf("%v\n", err)
	}

	for _, e := range all {
		fmt.Fprintf(w, "<item>\n")
		fmt.Fprintf(w, "<title>%s - %s</title>\n", html.EscapeString(e.Artist), html.EscapeString(e.Title))
		fmt.Fprintf(w, "<pubDate>%s</pubDate>\n", e.DateAdded)
		if e.Link != "" {
			fmt.Fprintf(w, "<link>%s</link>\n", html.EscapeString(e.Link))
		}
		fmt.Fprintf(w, "</item>\n")
	}

	fmt.Fprintf(w, "</channel>\n")
	fmt.Fprintf(w, "</rss>\n")

	return
}

func init() {
	flag.StringVar(&SuperSecret, "s", SuperSecret, "secret key")
	flag.StringVar(&ListenOn, "l", ListenOn, "listen [address]:port")
	flag.StringVar(&TxtFile, "f", TxtFile, "path to log file")
	flag.Parse()

	db, err := sql.Open("sqlite3", TxtFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		panic(err)
	}

	repo = NewSQLiteRepo(db)

	if err := repo.Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		panic(err)
	}
	if err := repo.Migrate(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		panic(err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			handlePost(w, r)
		case "GET":
			handleGet(w, r)
		default:
			// TODO return method not supported
			http.NotFound(w, r)
		}
		return
	})
	mux.HandleFunc("/rss.xml", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handleRss(w, r)
		default:
			// TODO return method not supported
			http.NotFound(w, r)
		}
		return
	})
	http.ListenAndServe(ListenOn, mux)
}
