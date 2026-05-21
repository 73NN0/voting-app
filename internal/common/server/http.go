package server

import (
	"net/http"

	"github.com/73NN0/voting-app/internal/common/logger"
)

// Middleware type
type Middleware func(http.Handler) http.Handler

// Chain middlewares (apply in order)
func Chain(handler http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

// Logging middleware
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logger.Logger.Info("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// Recovery middleware (panic handler)
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Logger.Error("panic: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORS middleware (simple, customisable)
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // or specific origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Router struct (custom mux with grouping)
type Router struct {
	mux *http.ServeMux
}

func NewRouter() *Router {
	return &Router{mux: http.NewServeMux()}
}

func (r *Router) Group(prefix string, fn func(*Router)) {
	sub := NewRouter()
	fn(sub)
	r.mux.Handle(prefix+"/", http.StripPrefix(prefix, sub.mux))
}

func (r *Router) Handle(pattern string, handler http.Handler, mws ...Middleware) {
	r.mux.Handle(pattern, Chain(handler, mws...))
}

func (r *Router) Handler() http.Handler {
	return r.mux
}
