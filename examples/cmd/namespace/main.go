package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"

	"github.com/ngergs/k8s-adm-ctrl/admissionreview"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var port = flag.Int("port", 8080, "Port on which the container listens for HTTP requests")
var tlsCrt = flag.String("tls_crt", "", "Path to the tls certificate")
var tlsPrivKey = flag.String("tls_priv_key", "", "Path to the tls private key")

func main() {
	setupHttpHandles()
	serveHttp()
}

// parseInputs read the TLS flags from command line and checks their consistency.
func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
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
// For this example we just provide a health endpoint and configure our example NamespaceLabelModifier.
func setupHttpHandles() {
	mutater := &namespaceLabelMutater{}
	// Adjust this to place your custom handlers
	http.Handle("/mutate", admissionreview.MutatingReviewer(mutater.Patch, compatibleGroupVersionKind))
	http.Handle("/validate", admissionreview.ValidatingReviewer(mutater.Validate, compatibleGroupVersionKind))
	http.HandleFunc("/health", handleHealthCheck)
}

// starts the HTTP server. TLS is activated if tlsCrt is set.
func serveHttp() {
	var err error
	if *tlsCrt != "" {
		log.Info().Msg("Serving HTTPS")
		err = http.ListenAndServeTLS(":"+strconv.Itoa(*port), *tlsCrt, *tlsPrivKey, nil)
	} else {
		log.Info().Msg("Serving HTTP")
		err = http.ListenAndServe(":"+strconv.Itoa(*port), nil)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("HTTP ListenAndServe failed")
	}
}
