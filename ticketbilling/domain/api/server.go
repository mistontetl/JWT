package api

import (
	"log"
	"net/http"
	"portal_autofacturacion/domain/api/handler"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Server struct {
	srv *http.Server
}

func New(port string, hdl *handler.BillingHistoryHandler) *Server {

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
	router.HandleFunc("/api/v1/verify", hdl.VerifyEstadioTicket).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/verify/customer", hdl.CustomerToVerifiedOrder).Methods(http.MethodPost)
	router.HandleFunc("/api/v1/verify/status", hdl.VerifyBillingStatus).Methods(http.MethodPost)

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
