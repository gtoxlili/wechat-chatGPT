package util

import (
	m "github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
	"net/http"
	"wxChatGPT/config"
)

func TodoEvent(w http.ResponseWriter) {
	_, err := w.Write([]byte{})
	if err != nil {
		log.Errorln(err)
		if config.GetIsDebug() {
			m.PrintPrettyStack(err)
		}
	}
}
