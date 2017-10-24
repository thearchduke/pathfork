package pathfork

import (
	"testing"
)

func TestURLFor(t *testing.T) {
	InitRoutes()
	u := URLFor("home")
	if u != "/" {
		t.Errorf("URLFor('home') returned %s, not root", u)
	}
}

func TestStaticURL(t *testing.T) {
	InitRoutes()
	u := StaticURL("css/main.css")
	if u != "/static/css/main.css" {
		t.Errorf("URLFor('css/main.css') returned %s, not /static/css/main.css", u)
	}
}

func TestNewTemplateRenderer(t *testing.T) {
	tr := NewTemplateRenderer()
	t.Log(tr)
}
