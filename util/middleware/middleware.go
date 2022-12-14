package middleware

import (
	m "github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
	"net/http"
	"runtime"
	"wxChatGPT/config"
	"wxChatGPT/util"
)

func Logger(next http.Handler) http.Handler {
	return m.RequestLogger(
		&m.DefaultLogFormatter{
			Logger:  log.StandardLogger(),
			NoColor: runtime.GOOS == "windows",
		})(next)
}

func Recover(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Errorln(err)
				if config.GetIsDebug() {
					m.PrintPrettyStack(err)
				}
				util.TodoEvent(w)
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
