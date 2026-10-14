package locale

import (
	"strings"
	"sync"

	"github.com/jeremygprawira/herr"
	"github.com/jeremygprawira/herr/localizer/mapl"
)

const Default = "id"

var (
	mu       sync.Mutex
	catalogs = map[string]map[string]string{} // lang -> code -> message
)

// Register merges a module's messages for one language. Later calls never clobber other modules.
func Register(lang string, msgs map[string]string) {
	mu.Lock()
	defer mu.Unlock()
	m, ok := catalogs[lang]
	if !ok {
		m = map[string]string{}
		catalogs[lang] = m
	}
	for code, msg := range msgs {
		if _, dup := m[code]; dup {
			panic("locale: duplicate message for " + lang + "/" + code)
		}
		m[code] = msg
	}
}

// Install hands every registered catalog to herr. Call once at boot, after all modules registered.
func Install() {
	mu.Lock()
	defer mu.Unlock()
	tables := make(map[string]map[string]string, len(catalogs))
	for lang, m := range catalogs {
		t := make(map[string]string, len(m))
		for code, msg := range m {
			t[Key(code)] = msg
		}
		tables[lang] = t
	}
	herr.SetLocalizer(mapl.New(tables))
}

// Key is herr's derived i18n key for a code.
func Key(code string) string {
	return "errors." + strings.ToLower(code) + ".message"
}

// Message returns the registered message for lang/code, if any.
func Message(lang, code string) (string, bool) {
	mu.Lock()
	defer mu.Unlock()
	msg, ok := catalogs[lang][code]
	return msg, ok
}

// Pick maps an Accept-Language header to "en" or "id". Anything else is "id".
func Pick(acceptLanguage string) string {
	for _, part := range strings.Split(acceptLanguage, ",") {
		tag := strings.TrimSpace(strings.SplitN(part, ";", 2)[0])
		tag = strings.ToLower(strings.SplitN(tag, "-", 2)[0])
		switch tag {
		case "en":
			return "en"
		case "id":
			return "id"
		}
	}
	return Default
}
