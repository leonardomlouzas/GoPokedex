package pokeClient

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetMap(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		expectedResponse := MapResponse{
			Count:    1,
			Next:     "next_url",
			Previous: "prev_url",
			Results: []APIResource{
				{Name: "location1", Url: "url1"},
			},
		}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(expectedResponse)
		}))
		defer server.Close()

		results, prev, next, err := GetMap(server.URL)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !reflect.DeepEqual(results, expectedResponse.Results) {
			t.Errorf("expected results %v, got %v", expectedResponse.Results, results)
		}
		if prev != expectedResponse.Previous {
			t.Errorf("expected previous %s, got %s", expectedResponse.Previous, prev)
		}
		if next != expectedResponse.Next {
			t.Errorf("expected next %s, got %s", expectedResponse.Next, next)
		}
	})

	t.Run("empty URL", func(t *testing.T) {
		_, _, _, err := GetMap("")
		if err == nil {
			t.Fatal("expected an error for empty URL, got nil")
		}
		expectedErrorMsg := "url is empty"
		if err.Error() != expectedErrorMsg {
			t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
		}
	})

	t.Run("non-OK status code", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		_, _, _, err := GetMap(server.URL)
		if err == nil {
			t.Fatal("expected an error for non-OK status, got nil")
		}
		expectedErrorMsg := fmt.Sprintf("error: received status code %d", http.StatusNotFound)
		if err.Error() != expectedErrorMsg {
			t.Errorf("expected error message '%s', got '%s'", expectedErrorMsg, err.Error())
		}
	})
}