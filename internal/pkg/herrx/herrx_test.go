package herrx

import (
	"testing"

	"github.com/jeremygprawira/herr"
)

func TestCatalogComplete(t *testing.T) { AssertCatalogComplete(t) }

func TestDuplicateCodePanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	Def("COMMON_NOT_FOUND", herr.KindNotFound, "again")
}

func TestDeprecatedIs410(t *testing.T) {
	if got := ErrDeprecatedEndpoint.New().HTTPStatus(); got != 410 {
		t.Fatalf("status = %d", got)
	}
}
