package application_test

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-checker/application"
	"url-checker/datamodel"
)

// This struct mocks and creates a map of DB objects to mimic the behaviour of a Database.
// From the perspective of the application layer we just need to cover cases when we do/dont
// find a resource in the DB (i.e. in the map of db objects, since we are testing the
// functionality of the GetMalware func.
type MockDb struct {
	malwareDB map[string]*datamodel.UrlModel
	err       error
}

// This Mock function checks the db object map to check if it consists of a db object with a URL equal
// to that of the input search Url string. Also, If the DB is uninitialized then we return nil
// as the db object map, which suggests that there were issues in instantiation of the database.
// If the db is initialized and if there is no match to the input search url in the map we return
// the unknown response to mimic store layer call of the corresponding GetMalware function.

// We are able to mimic store layer behaviour in constant time as map searches are O(1) operations
func (m *MockDb) GetInfo(searchUrl string) *datamodel.UrlModel {
	if m.malwareDB == nil {
		return nil
	}
	if malware, ok := m.malwareDB[searchUrl]; ok {
		return malware
	} else {
		return &datamodel.UrlModel{
			Url:      searchUrl,
			Risk:     "Unknown",
			Category: "Unknown",
		}
	}
}

func stringAddress(input string) *string {
	return &input
}

func TestApp_GetMalware(t *testing.T) {
	db := &MockDb{
		malwareDB: map[string]*datamodel.UrlModel{
			"http://www.compdata.ca/catalog":        {Url: "http://www.compdata.ca/catalog", Risk: "High", Category: "Malware"},
			"http://media0.mypage.cz/files/dc5.exe": {Url: "http://media0.mypage.cz/files/dc5.exe", Risk: "High", Category: "Malware"},
			"https://hybrid-analysis.com/sample":    {Url: "https://hybrid-analysis.com/sample", Risk: "High", Category: "Malware"},
		},
	}

	app := application.App{
		DataBase: db,
	}

	for _, tc := range []struct {
		name        string
		getUrl      string
		expected    *string
		paramValues map[string]string
		expErr      string
	}{
		{
			name:     "Valid Url",
			getUrl:   "/urlinfo/1/www.compdata.ca/catalog",
			expected: stringAddress(`{"RequestId":"1","Url":"http://www.compdata.ca/catalog","Risk":"High","Category":"Malware"}` + "\n"),
			paramValues: map[string]string{
				"request_id":        "1",
				"hostname_and_port": "www.compdata.ca",
				"original_path":     "catalog",
			},
		},
		{
			name:     "Valid Url - Url not present in DB (Unknown Link)",
			getUrl:   "/urlinfo/1/www.radioactive.co.uk/radio",
			expected: stringAddress(`{"RequestId":"1","Url":"http://www.radioactive.co.uk/radio","Risk":"Unknown","Category":"Unknown"}` + "\n"),
			paramValues: map[string]string{
				"request_id":        "1",
				"hostname_and_port": "www.radioactive.co.uk",
				"original_path":     "radio",
			},
		},
		{
			name:     "Valid Url with query paramters",
			getUrl:   "/urlinfo/1/hybrid-analysis.com/sample?scheme=https",
			expected: stringAddress(`{"RequestId":"1","Url":"https://hybrid-analysis.com/sample","Risk":"High","Category":"Malware"}` + "\n"),
			paramValues: map[string]string{
				"request_id":        "1",
				"hostname_and_port": "hybrid-analysis.com",
				"original_path":     "sample",
			},
		},
		{
			name:     "Valid Url with encoded path",
			getUrl:   "/urlinfo/1/media0.mypage.cz/files%2Fdc5.exe",
			expected: stringAddress(`{"RequestId":"1","Url":"http://media0.mypage.cz/files/dc5.exe","Risk":"High","Category":"Malware"}` + "\n"),
			paramValues: map[string]string{
				"request_id":        "1",
				"hostname_and_port": "media0.mypage.cz",
				"original_path":     "files%2Fdc5.exe",
			},
		},
		{
			name:   "Invalid Url - invalid hostname",
			getUrl: "/urlinfo/1/www.make\\invalid.compdata.ca/catalog",
			paramValues: map[string]string{
				"request_id":        "1",
				"hostname_and_port": "www.make\\invalid.compdata.ca",
				"original_path":     "catalog",
			},
			expErr: "{\"error\":\"Invalid URL\"}\n",
		},
		{
			name:   "Invalid Url - invalid encoded path",
			getUrl: "/urlinfo/1/www.compdata.ca/catalog%2Fresource",
			paramValues: map[string]string{
				"request_id":        "1",
				"hostname_and_port": "www.compdata.ca",
				"original_path":     "catalog%2resource",
			},
			expErr: "{\"error\":\"invalid URL escape \\\"%2r\\\"\"}\n",
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", tc.getUrl, nil)
			r = mux.SetURLVars(r, tc.paramValues)
			w := httptest.NewRecorder()

			app.GetInfo(w, r)

			if tc.expErr != "" {
				require.Equal(t, tc.expErr, w.Body.String())
			} else {
				if w.Code != http.StatusOK {
					t.Errorf("handler returned wrong status code: got %v want %v", w.Code, http.StatusOK)
				}

				if got := w.Body.String(); tc.expected != nil && got != *tc.expected {
					t.Errorf("handler returned unexpected body: got %v want %v", got, tc.expected)
				} else if tc.expected == nil {
					require.Equal(t, "null\n", got)
				}
			}
		})
	}
}

func TestApp_GetMalware_WithDBInitFail(t *testing.T) {
	// Creating an empty Database to mimic the case where connection with DB fails
	// and we have a nil DB object
	app := application.App{
		DataBase: &MockDb{
			malwareDB: nil,
		},
	}

	r, _ := http.NewRequest("GET", "/urlinfo/1/www.compdata.ca/catalog", nil)
	w := httptest.NewRecorder()

	app.GetInfo(w, r)

	// The response body in this case should be null
	require.Equal(t, "null\n", w.Body.String())
}
