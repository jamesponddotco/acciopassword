package main

import (
	"os"

	"git.sr.ht/~jamesponddotco/acciopassword/cmd/acopwctl/internal/app"
)

func main() {
	os.Exit(app.Run())
}
