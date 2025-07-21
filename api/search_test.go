package api

import "testing"

func TestSearchStopSanitiseSearch(t *testing.T) {
	verify := func(search string, output string) {
		got := sanitiseSeach(search)
		if got != output {
			t.Errorf("%s != %s, got %s", search, output, got)
		}
	}

	verify(`syd airport`, `syd* airport*`)
	verify(`international airport`, `international* airport*`)
}
