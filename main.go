package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
	"github.com/selfenergy/k8s-admission-ctrl/admissionreview"
)

var tlsCrt = flag.String("tls_crt", "", "Path to the tls certificate")
var tlsPrivKey = flag.String("tls_priv_key", "", "Path to the tls private key")

func main() {
	initLogging()
	setupHttpHandles()
	serveHttp()
}

// initLogging sets up some zerologging configutations.
func initLogging() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

// parseInputs read the TLS flags from command line and checks their consistency.
func init() {
	flag.Parse()
	if (*tlsCrt != "" && *tlsPrivKey == "") || (*tlsCrt == "" && *tlsPrivKey != "") {
		log.Fatal().Msg("Inconsistent configuration. Either specify both the tls-crt and tls-priv-key or neither.")
	}
}

// handleHealthCheck provides a simple health handle that always returns HTTP 200 for GET requests.
func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	w.WriteHeader(http.StatusOK)
}

// setupHttpHandles wires the relevant http handles together.
// For this example we just provide a health endpoint and configure our example NamespaceLabelModifier as root handler.
func setupHttpHandles() {
	mutater := &NamespaceLabelMutater{}
	// Adjust this to place your custom handlers
	http.HandleFunc("/mutate", admissionreview.ToHandelFunc(admissionreview.MutatingReviewer(mutater.Patch)))
	http.HandleFunc("/validate", admissionreview.ToHandelFunc(admissionreview.ValidatingReviewer(mutater.Validate)))
	http.HandleFunc("/health", handleHealthCheck)
}

// starts the HTTP server. TLS is activated if tlsCrt is set.
func serveHttp() {
	var err error
	if *tlsCrt != "" {
		log.Info().Msg("Serving HTTPS")
		err = http.ListenAndServeTLS(":8080", *tlsCrt, *tlsPrivKey, nil)
	} else {
		log.Info().Msg("Serving HTTP")
		err = http.ListenAndServe(":8080", nil)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("HTTP ListenAndServe failed")
	}
}
