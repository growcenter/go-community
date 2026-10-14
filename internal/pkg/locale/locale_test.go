package locale

import "testing"

func TestPick(t *testing.T) {
	cases := map[string]string{"": "id", "fr": "id", "en-US,en;q=0.9": "en", "id-ID": "id", "de, en": "en"}
	for in, want := range cases {
		if got := Pick(in); got != want {
			t.Errorf("Pick(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRegisterRejectsDuplicate(t *testing.T) {
	Register("xx", map[string]string{"A": "1"})
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate")
		}
	}()
	Register("xx", map[string]string{"A": "2"})
}

func TestKey(t *testing.T) {
	if got := Key("EVENT_QUOTA_FULL"); got != "errors.event_quota_full.message" {
		t.Fatal(got)
	}
}
