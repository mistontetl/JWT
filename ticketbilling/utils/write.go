package utils

import (
	"encoding/json"
	"net/http"
	"portal_autofacturacion/models"
)

func WriteAny(w http.ResponseWriter, response any) {
	w.Header().Add("Content-Type", "application/json")
	//	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
func WriteErr[T interface{}](w http.ResponseWriter, response T, err error, Code int) {
	if err != nil {
		w.WriteHeader(Code)
		WriteAny(w, models.ResponseServerModel[*int]{
			Code: Code, Msg: err.Error(),
		})
		return
	}
	WriteAny(w, models.ResponseServerModel[T]{
		Code: Code, Res: response,
	})
}
