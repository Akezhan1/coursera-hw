package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	InternalErrorQuery          = "fatal_query"
	TimeoutErrorQuery           = "timeout_query"
	InvalidToken                = "invalid_token"
	BadRequestErrorQuery        = "bad_request_query"
	BadRequestUnknownErrorQuery = "bad_request_unknown_query"
	InvalidJSONErrorQuery       = "invalid_json_query"
)

var rowsPool = sync.Pool{
	New: func() interface{} {
		return new(row)
	},
}

var usersPool = sync.Pool{
	New: func() interface{} {
		return new(User)
	},
}

type row struct {
	ID        int    `xml:"id"`
	Age       int    `xml:"age"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Gender    string `xml:"gender"`
	About     string `xml:"about"`
}

func match(r row, query string) bool {
	if strings.Contains(r.About, query) || strings.Contains(r.FirstName+" "+r.LastName, query) || query == "" {
		return true
	}
	return false
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")

	// check to errors
	if query == TimeoutErrorQuery {
		time.Sleep(time.Second * 2)
		return
	}
	if query == InternalErrorQuery {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if query == BadRequestErrorQuery {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if query == BadRequestUnknownErrorQuery {
		resp, _ := json.Marshal(SearchErrorResponse{"UnknownError"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}
	if query == InvalidJSONErrorQuery {
		w.Write([]byte("invalid_json"))
		return
	}

	if r.Header.Get("AccessToken") == InvalidToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	orderField := r.FormValue("order_field")
	orderBy, err := strconv.Atoi(r.FormValue("order_by"))
	if err != nil {
		panic(err)
	}
	if orderField == "" {
		orderField = "Name"
	}
	if (orderField != "Id" && orderField != "Age" && orderField != "Name") || (orderBy != 0 && orderBy != 1 && orderBy != -1) {
		resp, _ := json.Marshal(SearchErrorResponse{"ErrorBadOrderField"})
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
		return
	}

	// searching start
	f, err := os.Open("dataset.xml")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	decoder := xml.NewDecoder(f)
	users := make([]User, 0, 10)

	for {
		token, err := decoder.Token()
		if token == nil {
			break
		}
		if err != nil {
			panic(err)
		}

		switch tp := token.(type) {
		case xml.StartElement:
			if tp.Name.Local == "row" {
				row := rowsPool.Get().(*row)
				decoder.DecodeElement(&row, &tp)
				rowsPool.Put(row)
				if match(*row, query) {

					user := usersPool.Get().(*User)
					user.Id = row.ID
					user.Name = row.FirstName + " " + row.LastName
					user.Gender = row.Gender
					user.Age = row.Age
					user.About = row.About
					users = append(users, *user)
					usersPool.Put(user)
				}
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(users)
	//fmt.Println(string(data))
	w.Write(data)
}

func TestFindUserErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	type errorstestcases struct {
		Request       SearchRequest
		URL           string
		AccessToken   string
		ErrorExact    string
		ErrorContains string
	}

	cases := []errorstestcases{
		{
			URL:           "http://",
			ErrorContains: "unknown error",
		},
		{
			Request:       SearchRequest{Query: TimeoutErrorQuery},
			ErrorContains: "timeout for",
		},
		{
			AccessToken: InvalidToken,
			ErrorExact:  "Bad AccessToken",
		},
		{
			Request:    SearchRequest{Query: InternalErrorQuery},
			ErrorExact: "SearchServer fatal error",
		},
		{
			Request:       SearchRequest{Query: BadRequestErrorQuery},
			ErrorContains: "cant unpack error json",
		},
		{
			Request:       SearchRequest{Query: BadRequestUnknownErrorQuery},
			ErrorContains: "unknown bad request error",
		},
		{
			Request:    SearchRequest{OrderField: "order_field"},
			ErrorExact: "OrderFeld order_field invalid",
		},
		{
			Request:       SearchRequest{Query: InvalidJSONErrorQuery},
			ErrorContains: "cant unpack result json",
		},
	}

	for caseNum, item := range cases {
		url := server.URL
		if item.URL != "" {
			url = item.URL
		}

		client := SearchClient{
			URL:         url,
			AccessToken: item.AccessToken,
		}
		response, err := client.FindUsers(item.Request)

		if response != nil || err == nil {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}

		if item.ErrorExact != "" && err.Error() != item.ErrorExact {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, item.ErrorExact, err.Error())
		}

		if item.ErrorContains != "" && !strings.Contains(err.Error(), item.ErrorContains) {
			t.Errorf("[%d] wrong result, expected %#v to contain %#v", caseNum, err.Error(), item.ErrorContains)
		}
	}
}

func TestFindUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	type result struct {
		Response *SearchResponse
		Err      error
	}

	type testcases struct {
		Request SearchRequest
		Result  result
	}

	cases := []testcases{
		{
			Request: SearchRequest{
				Query:  "Boyd Wolf",
				Limit:  26,
				Offset: 0,
			},
			Result: result{
				Response: &SearchResponse{
					Users: []User{
						User{
							Id:     0,
							Name:   "Boyd Wolf",
							Age:    22,
							Gender: "male",
							About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
						},
					},
					NextPage: false,
				},
				Err: nil,
			},
		},
		{
			Request: SearchRequest{
				Query:   "J",
				Limit:   3,
				Offset:  0,
				OrderBy: 0,
			},
			Result: result{
				Response: &SearchResponse{
					Users: []User{
						User{
							Id:     6,
							Name:   "Jennings Mays",
							Age:    39,
							About:  "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
							Gender: "male",
						},
						User{
							Id:     8,
							Name:   "Glenn Jordan",
							Age:    29,
							About:  "Duis reprehenderit sit velit exercitation non aliqua magna quis ad excepteur anim. Eu cillum cupidatat sit magna cillum irure occaecat sunt officia officia deserunt irure. Cupidatat dolor cupidatat ipsum minim consequat Lorem adipisicing. Labore fugiat cupidatat nostrud voluptate ea eu pariatur non. Ipsum quis occaecat irure amet esse eu fugiat deserunt incididunt Lorem esse duis occaecat mollit.\n",
							Gender: "male",
						},
						User{
							Id:     21,
							Name:   "Johns Whitney",
							Age:    26,
							About:  "Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n",
							Gender: "male",
						},
					},
					NextPage: true,
				},
				Err: nil,
			},
		},
	}

	for _, item := range cases {
		s := &SearchClient{
			AccessToken: "token",
			URL:         server.URL,
		}
		result, err := s.FindUsers(item.Request)
		if err != nil || !reflect.DeepEqual(item.Result.Response, result) {
			t.Errorf("wrong result, \n\nexpected result: %#v\n\n got result: %#v,\n \nerror: %#v", item.Result.Response, result, err)
		}
	}
}

func TestOffsetLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer server.Close()

	type result struct {
		Response *SearchResponse
		Err      error
	}

	type testcases struct {
		Request SearchRequest
		Result  result
	}

	cases := []testcases{
		{
			Request: SearchRequest{
				Limit:      5,
				Offset:     -1,
				Query:      "Jennings",
				OrderField: "",
				OrderBy:    OrderByAsc,
			},
			Result: result{
				Response: nil,
				Err:      errors.New("offset must be > 0"),
			},
		},
		{
			Request: SearchRequest{
				Limit:      -5,
				Offset:     1,
				Query:      "",
				OrderField: "",
				OrderBy:    OrderByAsc,
			},
			Result: result{
				Response: nil,
				Err:      errors.New("limit must be > 0"),
			},
		},
	}

	for _, item := range cases {
		s := &SearchClient{
			AccessToken: "token",
			URL:         server.URL,
		}
		r, err := s.FindUsers(item.Request)
		if r != nil || err.Error() != item.Result.Err.Error() {
			t.Errorf("wrong result: \n\nexpected \nresult: %#v\n error: %#v,\n\n got \nresult: %#v, \n error: %#v", item.Result.Response, item.Result.Err, r, err)
		}
	}

}
