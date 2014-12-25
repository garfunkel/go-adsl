package adsl

import (
	"log"
	"testing"
)

func TestLookup(t *testing.T) {
	_, err := Lookup("59/47 Hampstead Road, Homebush West, NSW 2140")

	if err != nil {
		log.Fatal(err)
	}
}
