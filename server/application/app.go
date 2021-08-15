package application

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"url-checker/datastore"
)

const (
	URLDelimiter = "/"
	SchemeSuffix = "://"
)

type App struct {
	DataBase datastore.DB
	Router   *mux.Router
	Server   *http.Server
}

type Response struct {
	RequestId string
	Url       string
	Risk      string
	Category  string
}

func NewApp(db datastore.DB) App {
	router := mux.NewRouter().StrictSlash(true).UseEncodedPath()
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	app := App{
		DataBase: db,
		Router:   router,
		Server:   server,
	}
	app.initializeRoutes()
	return app
}

func (a *App) Serve() error {
	log.Println("Web server is available on port 8080.")
	return a.Server.ListenAndServe()
}

func (a *App) initializeRoutes() {
	path := "/urlinfo/{request_id}/{hostname_and_port}/{original_path}"
	a.Router.HandleFunc(path, a.GetMalware).Methods(http.MethodGet)
	a.Router.HandleFunc(path, a.GetMalware).Queries("scheme", "{scheme}").Methods(http.MethodGet)
}

func (a *App) GetMalware(w http.ResponseWriter, r *http.Request) {
	var searchUrl string
	var scheme string
	var response *Response
	w.Header().Set("Content-Type", "application/json")

	// Fetching parameters
	params := mux.Vars(r)
	requestId := params["request_id"]
	hostName := params["hostname_and_port"]
	originalPath, err := url.PathUnescape(params["original_path"])
	if err != nil {
		log.Println("Error decoding resource original path")
		sendErr(w, http.StatusBadRequest, err.Error())
		return
	}

	// retrieving the scheme from the query parameters, default is set to http
	err = r.ParseForm()
	if err != nil {
		log.Println("Error parsing query params.")
		sendErr(w, http.StatusBadRequest, err.Error())
		return
	}
	schemes := r.Form["scheme"]
	if len(schemes) == 0 {
		scheme = "http"
	} else {
		scheme = schemes[0]
	}
	scheme += SchemeSuffix

	// Form the search URL for the scheme
	searchUrl = scheme + hostName + URLDelimiter + originalPath

	// validate the URL, Sends a bad request error if invalid search URL
	_, err = url.ParseRequestURI(searchUrl)
	if err != nil {
		sendErr(w, http.StatusBadRequest, "Invalid URL")
		return
	}

	// Fetching the malware info for the search url
	malware := a.DataBase.GetMalware(searchUrl)
	if malware != nil {
		response = &Response{
			RequestId: requestId,
			Url:       malware.Url,
			Risk:      malware.Risk,
			Category:  malware.Category,
		}
	}

	// Encoding the response
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		sendErr(w, http.StatusInternalServerError, err.Error())
	}
}

func sendErr(w http.ResponseWriter, code int, message string) {
	resp, _ := json.Marshal(map[string]string{"error": message})
	http.Error(w, string(resp), code)
}
