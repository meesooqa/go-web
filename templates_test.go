package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoadTemplates_Success(t *testing.T) {
	tmpl, err := LoadTemplates("testdata/templates")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tmpl == nil {
		t.Fatal("LoadTemplates returned nil without an error")
	}
}

func TestLoadTemplates_MissingDir(t *testing.T) {
	_, err := LoadTemplates("testdata/does_not_exist")
	if err == nil {
		t.Fatal("expected an error for a non-existent directory")
	}
}

func TestLoadTemplates_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadTemplates(dir)
	if err == nil {
		t.Fatal("expected an error if the directory has no template files")
	}
}

func TestTemplates_Render(t *testing.T) {
	tmpl, err := LoadTemplates("testdata/templates")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	data := map[string]any{"Title": "Hello", "Message": "world"}

	if err := tmpl.Render(rec, 200, "index.html", data); err != nil {
		t.Fatalf("unexpected rendering error: %v", err)
	}

	if rec.Code != 200 {
		t.Errorf("status = %d, expected 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, expected text/html", ct)
	}
	if !strings.Contains(rec.Body.String(), "Hello") {
		t.Errorf("response body does not contain the expected header: %s", rec.Body.String())
	}
}

func TestTemplates_Render_EscapesHTML(t *testing.T) {
	tmpl, err := LoadTemplates("testdata/templates")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	data := map[string]any{"Title": "<script>alert(1)</script>", "Message": "ok"}

	if err := tmpl.Render(rec, 200, "index.html", data); err != nil {
		t.Fatalf("unexpected rendering error: %v", err)
	}

	if strings.Contains(rec.Body.String(), "<script>alert(1)</script>") {
		t.Error("script tag was not escaped — potential XSS vulnerability")
	}
	if !strings.Contains(rec.Body.String(), "&lt;script&gt;") {
		t.Error("expected the escaped representation &lt;script&gt;")
	}
}

func TestTemplates_Render_ErrorDoesNotWritePartialResponse(t *testing.T) {
	tmpl, err := LoadTemplates("testdata/templates")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rec := httptest.NewRecorder()
	err = tmpl.Render(rec, 200, "broken.html", map[string]any{})
	if err == nil {
		t.Fatal("expected an error rendering the broken.html template")
	}

	// in the event of a rendering error, WriteHeader must not be
	// called at all — httptest.ResponseRecorder defaults to Code == 200,
	// so we verify that the body remains empty (meaning the write
	// never actually reached the ResponseWriter).
	if rec.Body.Len() != 0 {
		t.Errorf("response body must be empty on rendering error, got: %q", rec.Body.String())
	}
}
