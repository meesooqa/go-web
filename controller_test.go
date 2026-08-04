package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type multiRouteController struct{}

func (multiRouteController) Routes() []Route {
	return []Route{
		{Pattern: "GET /a", Handler: func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("a")) }},
		{Pattern: "POST /a", Handler: func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("post-a")) }},
	}
}

func TestServer_Register_MethodSpecificRouting(t *testing.T) {
	srv, err := New(testConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	srv.Register(multiRouteController{})

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	getResp, err := http.Get(ts.URL + "/a")
	if err != nil {
		t.Fatalf("GET request error: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Errorf("GET /a: status = %d, expected 200", getResp.StatusCode)
	}

	postResp, err := http.Post(ts.URL+"/a", "text/plain", nil)
	if err != nil {
		t.Fatalf("POST request error: %v", err)
	}
	defer postResp.Body.Close()
	if postResp.StatusCode != http.StatusOK {
		t.Errorf("POST /a: status = %d, expected 200", postResp.StatusCode)
	}

	deleteReq, _ := http.NewRequest(http.MethodDelete, ts.URL+"/a", nil)
	deleteResp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		t.Fatalf("DELETE request error: %v", err)
	}
	defer deleteResp.Body.Close()
	if deleteResp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("DELETE /a: status = %d, expected 405 (route registered only for GET/POST)", deleteResp.StatusCode)
	}
}

func TestServer_Register_MultipleControllers(t *testing.T) {
	srv, err := New(testConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctrlA := routeSet{"GET /one": "one"}
	ctrlB := routeSet{"GET /two": "two"}
	srv.Register(ctrlA, ctrlB)

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	for _, path := range []string{"/one", "/two"} {
		resp, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatalf("request error %s: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: status = %d, expected 200", path, resp.StatusCode)
		}
	}
}

type routeSet map[string]string

func (rs routeSet) Routes() []Route {
	var routes []Route
	for pattern, body := range rs {
		body := body
		routes = append(routes, Route{
			Pattern: pattern,
			Handler: func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(body)) },
		})
	}
	return routes
}
