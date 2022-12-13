package middleware

import (
	m "github.com/go-chi/chi/v5/middleware"
	log "github.com/sirupsen/logrus"
	"net/http"
	"runtime"
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
				if r.URL.Path == "/healthCheck" {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("error"))
					m.PrintPrettyStack(err)
				} else {
					log.Errorln(err)
					util.TodoEvent(w)
				}
			}
		}()
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
