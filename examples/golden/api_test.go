package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	api "example.com/golden"
	"example.com/golden/golden"
	"github.com/google/go-cmp/cmp"
)

func TestGreetAPI(t *testing.T) {
	a := api.NewAPI()
	router := api.NewRouter(a)

	srv := httptest.NewServer(router)
	defer srv.Close()

	client := srv.Client()
	resp, err := client.Post(srv.URL+"/greet", "application/json", golden.Open(t, "request.json"))
	if err != nil {
		t.Fatalf("make request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	assertResponse(t, resp, "response.json")
}

func assertResponse(t *testing.T, resp *http.Response, fixturePath string) {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	assertEqualJSON(t, golden.ReadBytes(t, fixturePath), body)
}

func assertEqualJSON(t *testing.T, wantJSON, gotJSON []byte) {
	t.Helper()

	var want, got any
	if err := json.Unmarshal(wantJSON, &want); err != nil {
		t.Fatalf("unmarshal want JSON: %v", err)
	}
	if err := json.Unmarshal(gotJSON, &got); err != nil {
		t.Fatalf("unmarshal got JSON: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("JSON mismatch (-want +got):\n%s", diff)
	}
}
