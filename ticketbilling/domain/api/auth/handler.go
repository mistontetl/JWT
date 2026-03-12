package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"portal_autofacturacion/models"
	"portal_autofacturacion/utils"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErr(w, "", err, http.StatusBadRequest)
		return
	}
	if err := h.svc.Register(req.Username, req.Password); err != nil {
		utils.WriteErr(w, "", err, http.StatusBadRequest)
		return
	}
	utils.WriteAny(w, models.ResponseServerModel[any]{Code: http.StatusCreated, Msg: "usuario creado", Datetime: utils.DateTime()})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErr(w, "", err, http.StatusBadRequest)
		return
	}
	res, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		utils.WriteErr(w, "", err, http.StatusUnauthorized)
		return
	}
	utils.WriteAny(w, models.ResponseServerModel[any]{Code: http.StatusOK, Msg: "ok", Res: res, Datetime: utils.DateTime()})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	if token == "" {
		utils.WriteErr(w, "", http.ErrNoCookie, http.StatusUnauthorized)
		return
	}
	if err := h.svc.Logout(token); err != nil {
		utils.WriteErr(w, "", err, http.StatusUnauthorized)
		return
	}
	utils.WriteAny(w, models.ResponseServerModel[any]{Code: http.StatusOK, Msg: "sesión cerrada", Datetime: utils.DateTime()})
}

func (h *Handler) SessionInfo(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
	claims, err := h.svc.ValidateAndTouch(token)
	if err != nil {
		utils.WriteErr(w, "", err, http.StatusUnauthorized)
		return
	}
	utils.WriteAny(w, models.ResponseServerModel[any]{
		Code:     http.StatusOK,
		Msg:      "sesión activa",
		Datetime: utils.DateTime(),
		Res: map[string]any{
			"username": claims.Subject,
			"session":  claims.SessionID,
		},
	})
}

func (h *Handler) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if token == "" {
			utils.WriteErr(w, "", http.ErrNoCookie, http.StatusUnauthorized)
			return
		}
		if _, err := h.svc.ValidateAndTouch(token); err != nil {
			utils.WriteErr(w, "", err, http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
