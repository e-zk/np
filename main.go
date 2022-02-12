package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var (
	Bullet      = "&bullet;"
	SuperSecret = "super secret key"
	ListenOn    = ":8081"
	TxtFile     = "./nowplaying.txt"
)

func handlePost(w http.ResponseWriter, r *http.Request) {
	var (
		artist = r.PostFormValue("artist")
		track  = r.PostFormValue("track")
		album  = r.PostFormValue("album")
		url    = r.PostFormValue("url")
		secret = r.PostFormValue("key")
	)

	if secret != SuperSecret {
		http.Error(w, "you are forbidden.", http.StatusBadRequest)
		return
	}

	// album is the same as a track internally
	if album != "" {
		track = album
	}

	if artist == "" || track == "" {
		http.Error(w, "artist and either track or album must be specified.", http.StatusBadRequest)
		return
	}

	f, err := os.OpenFile(TxtFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		http.Error(w, "could not open txt file.", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "%s - %s", artist, track)

	if url != "" {
		fmt.Fprintf(f, "\t%s", url)
	}
	fmt.Fprintf(f, "\n")

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

	// open text file
	f, err := os.Open(TxtFile)
	if err != nil {
		http.Error(w, "could not open file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	fs := bufio.NewScanner(f)

	if !plaintext {
		fmt.Fprintf(w, "<pre>\n")
	}

	// print each line
	for fs.Scan() {
		txt := strings.Split(fs.Text(), "\t")
		if plaintext {
			fmt.Fprintf(w, "%s\n", txt[0])
		} else if len(txt) == 1 {
			fmt.Fprintf(w, "%s %s\n", Bullet, txt[0])
		} else {
			fmt.Fprintf(w, "%s %s [<a href=\"%s\">&Hat;</a>]\n", Bullet, txt[0], txt[1])
		}
	}

	if !plaintext {
		fmt.Fprintf(w, "</pre>\n")
	}

	return
}

func handleRss(w http.ResponseWriter, r *http.Request) {
	// open text file
	f, err := os.Open(TxtFile)
	if err != nil {
		http.Error(w, "could not open file", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	fs := bufio.NewScanner(f)

	w.Header().Set("Content-Type", "application/xml")

	fmt.Fprintf(w, "<rss version=\"2.0\">\n")
	fmt.Fprintf(w, "<channel>\n")
	fmt.Fprintf(w, "<title>music log</title>\n")
	fmt.Fprintf(w, "<description/>\n")

	// print each line
	for fs.Scan() {
		txt := strings.Split(fs.Text(), "\t")
		fmt.Fprintf(w, "<item>\n")
		fmt.Fprintf(w, "<title>%s</title>\n", txt[0])
		if len(txt) != 1 {
			fmt.Fprintf(w, "<link>%s</link>\n", txt[1])
		}
		fmt.Fprintf(w, "</item>\n")
	}

	fmt.Fprintf(w, "</channel>\n")
	fmt.Fprintf(w, "</rss>\n")

	return
}

func main() {
	flag.StringVar(&SuperSecret, "s", SuperSecret, "secret key")
	flag.StringVar(&ListenOn, "l", ListenOn, "listen [address]:port")
	flag.StringVar(&TxtFile, "f", TxtFile, "path to log file")
	flag.Parse()

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
