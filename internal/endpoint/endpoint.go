package endpoint

import "git.sr.ht/~jamesponddotco/acciopassword/internal/build"

const (
	// Root is the endpoint for the root handler.
	Root string = "/"

	// Random is the endpoint for the Random handler.
	Random string = Root + build.APIVersion + "/random/"

	// Diceware is the endpoint for the Diceware handler.
	Diceware string = Root + build.APIVersion + "/diceware/"

	// PIN is the endpoint for the PIN handler.
	PIN string = Root + build.APIVersion + "/pin/"

	// Metrics is the endpoint for the Metrics handler.
	Metrics string = Root + build.APIVersion + "/metrics/"

	// Health is the endpoint for the Health handler.
	Health string = Root + build.APIVersion + "/health/"

	// Ping is the endpoint for the Ping handler.
	Ping string = Root + build.APIVersion + "/ping/"
)
