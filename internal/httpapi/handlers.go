package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/lkjin41/orders-kafka/internal/orders"
)

type API struct {
	orders *orders.Repo
}

func New(ordersRepo *orders.Repo) *API {
	return &API{orders: ordersRepo}
}

func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("/orders", a.handleOrders)
}

func (a *API) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.createOrder(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) createOrder(w http.ResponseWriter, r *http.Request) {
	var req orders.CreateOrderRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	o, err := a.orders.Create(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(o)
}
