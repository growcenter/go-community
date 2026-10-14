package herrx

import (
	"testing"

	"github.com/jeremygprawira/herr"

	"go-community/internal/pkg/locale"
)

var (
	AllClasses      []*herr.Class
	EnglishMessages = map[string]string{}
)

// Def registers a catalog class. The HTTP status comes from kind.
func Def(code string, kind herr.Kind, en string) *herr.Class {
	return DefStatus(code, kind, 0, en)
}

// DefStatus is Def with an explicit HTTP status, for statuses herr has no Kind for (410).
func DefStatus(code string, kind herr.Kind, status int, en string) *herr.Class {
	if _, dup := EnglishMessages[code]; dup {
		panic("herrx: duplicate error code " + code)
	}
	c := herr.Define(herr.Class{Code: code, Kind: kind, HTTP: status, Public: herr.Public{Message: en}})
	EnglishMessages[code] = en
	AllClasses = append(AllClasses, c)
	return c
}

// AssertCatalogComplete fails the test when any class lacks an Indonesian message.
// Every module's test file calls it after importing its own catalog.
func AssertCatalogComplete(t *testing.T) {
	t.Helper()
	for _, c := range AllClasses {
		id, ok := locale.Message(locale.Default, c.Code)
		if !ok || id == "" {
			t.Errorf("%s has no %s message", c.Code, locale.Default)
			continue
		}
		if id == EnglishMessages[c.Code] {
			t.Errorf("%s: %s message equals English", c.Code, locale.Default)
		}
	}
}
