package api

import (
	"log"
	"net/http"
	"portal_autofacturacion/domain/api/handler"
	"time"

	"portal_autofacturacion/domain/api/auth"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Server struct {
	srv *http.Server
}

func New(port string, hdl *handler.BillingHistoryHandler, authHandler *auth.Handler) *Server {

	router := mux.NewRouter().StrictSlash(true)
	///

	///
	handler := cors.New(cors.Options{
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		AllowedOrigins:   []string{"http://*", "https://*"},
	}).Handler(router)

	// Routes
	router.HandleFunc("/api/v1/invoice", hdl.CreateBillingRequest).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/status", hdl.TrackingBillingRequest).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/verify/status", hdl.VerifyBillingStatus).Methods(http.MethodPost)
	///
	router.HandleFunc("/api/v1/auth/login", authHandler.Login).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/auth/logout", authHandler.Logout).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/auth/session", authHandler.SessionInfo).Methods(http.MethodGet)
	router.HandleFunc("/api/v1/auth/register", authHandler.Register).Methods(http.MethodPost)
	return &Server{
		srv: &http.Server{
			Addr:         ":" + port,
			Handler:      handler, //cors.Default().Handler(router),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	log.Printf("✅ Server running on http://localhost%s", s.srv.Addr)
	return s.srv.ListenAndServe()
}
