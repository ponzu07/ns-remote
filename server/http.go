package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"ns-remote/stream"

	"github.com/gorilla/mux"
	"github.com/mzyy94/nscon"
	"github.com/pion/webrtc/v2"
)

var mSource *stream.MediaSource

func webRTCOfferHandler(w http.ResponseWriter, r *http.Request) {
	var offer webrtc.SessionDescription
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "{\"error\": \"%s\"}", err.Error())
		return
	}

	if mSource.IsLinked {
		mSource.Unlink()
	}

	mStreamer := &stream.WebRTCStreamer{}
	answer, err := mStreamer.Setup(offer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "{\"error\": \"%s\"}", err.Error())
		return
	}

	mSource.Link(mStreamer)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&answer)
}

// StartHTTPServer starts HTTP server
func StartHTTPServer(mediaSource *stream.MediaSource, con *nscon.Controller) {
	mSource = mediaSource
	controller = con

	r := mux.NewRouter()

	r.HandleFunc("/connect", webRTCOfferHandler).Methods("POST")
	r.HandleFunc("/controller", controllerHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./")))
	http.Handle("/", r)

	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8000",
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
