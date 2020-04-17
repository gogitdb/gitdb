package gitdb_test

import (
	"net/http"
	"testing"
)

func TestServer(t *testing.T) {
	cfg := getConfig()
	cfg.EnableUI = true
	teardown := setup(t, cfg)

	insert(getTestMessageWithId(1), false)

	//fire off some requests
	client := http.DefaultClient
	requests := []*http.Request{
		request(http.MethodGet, "http://localhost:4120/css/app.css"),
		request(http.MethodGet, "http://localhost:4120/js/app.js"),
		request(http.MethodGet, "http://localhost:4120/"),
		request(http.MethodGet, "http://localhost:4120/errors/Message"),
		request(http.MethodGet, "http://localhost:4120/list/Message"),
		request(http.MethodGet, "http://localhost:4120/view/Message"),
		request(http.MethodGet, "http://localhost:4120/view/Message/b0/r0"),
	}

	for _, req := range requests {
		t.Logf("Testing %s", req.URL.String())
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("GitDB UI Server request failed: %s", err)
		}

		//todo use golden files to check response
		// b, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		// t.Log(string(b))
	}

	teardown(t)
}

func request(method, url string) *http.Request {
	req, _ := http.NewRequest(method, url, nil)
	return req
}
