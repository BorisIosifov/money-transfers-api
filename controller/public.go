package controller

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BorisIosifov/money-transfers-api/model"
)

// swagger:route GET /public/rates Public GetPublicRates
//
// Returns Rates
//
//	Responses:
//	  default: errorResult
//	  200: Rates
//	  400: errorResult
//	  500: errorResult
func (ctrl Controller) GetPublicRates(w http.ResponseWriter, r *http.Request) {
	rates, err := model.GetCurrentRates(ctrl.DB)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}

	resJSON, err := json.Marshal(rates)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ctrl.PrintError(w, r, fmt.Errorf("Internal server error: %s", err))
		return
	}
	fmt.Fprintln(w, string(resJSON))
}
