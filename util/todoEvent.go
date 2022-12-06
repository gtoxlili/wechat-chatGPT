package util

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

func TodoEvent(w http.ResponseWriter) {
	_, err := w.Write([]byte{})
	if err != nil {
		log.Errorln(err)
	}
}
