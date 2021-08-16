package main_test

import (
	"bytes"
	"context"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"url-checker/application"
	"url-checker/datastore"
)

func setupDatabaseAndCache() (*mongo.Client, *redis.Client, error) {
	dbClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, nil, err
	}
	cacheClient := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	return dbClient, cacheClient, nil
}

// This integration test is required for verifying the behaviour of server API routing
// and database querying and to see how components like application servers, cache, and
// database work in tandem.
func TestApp_Integration(t *testing.T) {

	if testing.Short() {
		t.Skip("Skipping this for Unit Tests")
	}

	dbClient, cacheClient, err := setupDatabaseAndCache()
	if dbClient != nil && cacheClient != nil {
		defer cacheClient.Close()
		defer dbClient.Disconnect(context.TODO())
	} else {
		if err != nil {
			t.Fatal("Database connectivity issues: ", err.Error())
		} else {
			t.Fatal("Cache connectivity issues.")
		}
	}

	store := datastore.NewMongo(dbClient, cacheClient)

	// creating an application object with a new Router and a Server
	app := application.NewApp(store)

	// Creating buffer to store the log outputs
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	for _, tc := range []struct {
		name     string
		getUrl   string
		expected string
		expErr   string
		cacheLog string
	}{
		{
			name:     "Valid Url",
			getUrl:   "/urlinfo/1/www.compdata.ca/catalog",
			expected: `{"RequestId":"1","Url":"http://www.compdata.ca/catalog","Risk":"High","Category":"Malware"}` + "\n",
		},
		{
			name:     "Valid Url - Running again to fetch from cache",
			getUrl:   "/urlinfo/1/www.compdata.ca/catalog",
			expected: `{"RequestId":"1","Url":"http://www.compdata.ca/catalog","Risk":"High","Category":"Malware"}` + "\n",
			cacheLog: "Returning Cached Results.",
		},
		{
			name:     "Valid Url - Unknown link",
			getUrl:   "/urlinfo/1/google.co.uk/catalog",
			expected: `{"RequestId":"1","Url":"http://google.co.uk/catalog","Risk":"Unknown","Category":"Unknown"}` + "\n",
		},
		{
			name:     "Valid Url with query paramters",
			getUrl:   "/urlinfo/1/hybrid-analysis.com/sample?scheme=https",
			expected: `{"RequestId":"1","Url":"https://hybrid-analysis.com/sample","Risk":"High","Category":"Malware"}` + "\n",
		},
		{
			name:     "Valid Url with encoded path",
			getUrl:   "/urlinfo/1/media0.mypage.cz/files%2Fdc5.exe",
			expected: `{"RequestId":"1","Url":"http://media0.mypage.cz/files/dc5.exe","Risk":"High","Category":"Malware"}` + "\n",
		},
		{
			name:   "Invalid Url - invalid hostname",
			getUrl: "/urlinfo/1/www.make\\invalid.compdata.ca/catalog",
			expErr: "{\"error\":\"Invalid URL\"}\n",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// Create a request 'r' to test our server and 'w' to capture the response
			r, _ := http.NewRequest("GET", tc.getUrl, nil)
			w := httptest.NewRecorder()

			// Starting the server and deferring the graceful shut down
			app.Router.ServeHTTP(w, r)
			defer app.Server.Shutdown(context.TODO())

			if tc.expErr != "" {
				require.Equal(t, tc.expErr, w.Body.String())
			} else {
				if w.Code != http.StatusOK {
					t.Errorf("handler returned wrong status code: got %v want %v", w.Code, http.StatusOK)
				}

				if got := w.Body.String(); got != tc.expected {
					t.Errorf("handler returned unexpected body: got %v want %v", got, tc.expected)
				}

				if tc.cacheLog != "" {
					require.Contains(t, buf.String(), tc.cacheLog)
				}
			}
		})
	}
}
