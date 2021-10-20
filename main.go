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

func main() {
	initLogging()
	tlsCrt, tlsPrivKey := parseInputs()
	setupHttpHandles()
	serveHttp(tlsCrt, tlsPrivKey)
}

// initLogging sets up some zerologging configutations.
func initLogging() {
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

// parseInputs read the TLS flags from command line and checks their consistency.
func parseInputs() (tlsCrt *string, tlsPrivKey *string) {
	tlsCrt = flag.String("tls_crt", "", "Path to the tls certificate")
	tlsPrivKey = flag.String("tls_priv_key", "", "Path to the tls private key")
	flag.Parse()
	if (*tlsCrt != "" && *tlsPrivKey == "") || (*tlsCrt == "" && *tlsPrivKey != "") {
		log.Fatal().Msg("Inconsistent configuration. Either specify both the tls-crt and tls-priv-key or neither.")
	}
	return
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
	// Adjust this to place your custom handlers
	http.HandleFunc("/mutate", admissionreview.ToHandler(
		&admissionreview.MutatingReviewer{
			Mutater: &NamespaceLabelMutater{},
		}))
	http.HandleFunc("/validate", admissionreview.ToHandler(
		&admissionreview.ValidatingReviewer{
			Validator: &NamespaceLabelMutater{},
		}))
	http.HandleFunc("/health", handleHealthCheck)
}

// starts the HTTP server. TLS is activated if tlsCrt is set.
func serveHttp(tlsCrt *string, tlsPrivKey *string) {
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
