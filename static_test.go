package main

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"regexp"
	"strings"
	"testing"
)

const maxRedirects = 3

// publicExtensions are the only kinds of file that belong in the web root.
var publicExtensions = map[string]bool{
	".html": true,
	".css":  true,
	".js":   true,
	".png":  true,
	".jpg":  true,
}

// getFollowingRedirects fetches target and follows the redirects that
// net/http adds for cleaned paths and index.html.
func getFollowingRedirects(h http.Handler, target string) *httptest.ResponseRecorder {
	current := target
	rec := do(h, http.MethodGet, current, "")
	for range maxRedirects {
		location := rec.Header().Get("Location")
		if rec.Code < http.StatusMovedPermanently || rec.Code > http.StatusPermanentRedirect || location == "" {
			break
		}

		base, _ := url.Parse(current)
		next, err := base.Parse(location)
		if err != nil {
			break
		}
		current = next.RequestURI()
		rec = do(h, http.MethodGet, current, "")
	}

	return rec
}

func TestIndexIsServed(t *testing.T) {
	h := newTestHandler(t, "http://irma.invalid")

	for _, method := range []string{http.MethodGet, http.MethodHead} {
		rec := do(h, method, "/", "")

		if rec.Code != http.StatusOK {
			t.Errorf("%s / status = %d, want 200", method, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
			t.Errorf("%s / Content-Type = %q, want text/html", method, ct)
		}
	}
	if body := do(h, http.MethodGet, "/", "").Body.String(); !strings.Contains(body, "YiviTube") {
		t.Errorf("GET / body is not the YiviTube page")
	}
}

func TestPublicAssetsAreServed(t *testing.T) {
	h := newTestHandler(t, "http://irma.invalid")

	for _, target := range []string{
		"/css/irmatube.css",
		"/js/mosaic.1.0.1.min.js",
		"/img/YiviTube_logo.png",
		"/content/movies.js",
		"/content/covers/up.jpg",
	} {
		if rec := do(h, http.MethodGet, target, ""); rec.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want 200", target, rec.Code)
		}
	}
}

func TestNonPublicFilesNotServed(t *testing.T) {
	h := newTestHandler(t, "http://irma.invalid")

	for _, target := range []string{
		"/.git/config",
		"/.git/HEAD",
		"/.github/workflows/delivery.yml",
		"/.gitignore",
		"/vendor/autoload.php",
		"/vendor/",
		"/tests/verifysession_test.php",
		"/tests/",
		"/node_modules/bootstrap/package.json",
		"/node_modules/",
		"/package.json",
		"/Dockerfile",
		"/docker-compose.yml",
		"/go.mod",
		"/go.sum",
		"/main.go",
		"/server.go",
		"/verifier.go",
		"/verify_test.go",
		"/CLAUDE.md",
		"/README.md",
		"/LICENSE",
		"/data/demo-privkey.pem",
		"/data/demo-publickey.pem",
		"/videos/up.json",
		"/videos/",
		"/static/index.html",
		"/php/session.php",
		"/php/verifysession.php",
		"/composer.json",
		"/config.php",
		"/index.php",
		"/index.html/../go.mod",
		"/css/../go.mod",
		"/%2e%2e/go.mod",
		"/..%2fgo.mod",
	} {
		rec := getFollowingRedirects(h, target)

		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s status = %d, want 404", target, rec.Code)
		}
	}
}

func TestDirectoriesAreNotListed(t *testing.T) {
	h := newTestHandler(t, "http://irma.invalid")

	for _, target := range []string{"/css/", "/css", "/img/", "/content/", "/content/covers/", "/lib/", "/js/"} {
		rec := do(h, http.MethodGet, target, "")

		if rec.Code == http.StatusOK {
			t.Errorf("GET %s status = 200, want no directory listing (body %q)", target, rec.Body.String())
		}
	}
}

func TestStaticFilesArePublic(t *testing.T) {
	static, err := fs.Sub(staticFiles, "static")
	if err != nil {
		t.Fatal(err)
	}

	err = fs.WalkDir(static, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !publicExtensions[path.Ext(name)] {
			t.Errorf("static/%s is embedded and would be served; allowed extensions are html, css, js, png and jpg", name)
		}

		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestVideosMatchMovieList(t *testing.T) {
	videos, err := loadVideos(videoFiles)
	if err != nil {
		t.Fatal(err)
	}
	movies, err := fs.ReadFile(staticFiles, "static/content/movies.js")
	if err != nil {
		t.Fatal(err)
	}

	listed := regexp.MustCompile(`id:\s*"([^"]+)"`).FindAllStringSubmatch(string(movies), -1)
	if len(listed) != len(videos) {
		t.Errorf("movies.js lists %d movies, videos/ has %d", len(listed), len(videos))
	}

	for _, m := range listed {
		v, ok := videos[m[1]]
		if !ok {
			t.Errorf("movies.js lists %q, which has no videos/%s.json", m[1], m[1])
			continue
		}
		if v.YoutubeID == "" {
			t.Errorf("%s has no youtubeId", m[1])
		}
	}
}
