package db_test

import (
	"os"
	"testing"

	"github.com/mitch-jensen/mymovies/internal/testdb"
)

func TestMain(m *testing.M) {
	os.Exit(testdb.Run(m))
}
