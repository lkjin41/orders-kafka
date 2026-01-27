package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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
	mux.HandleFunc("/orders/", a.handleOrderByID)

	mux.Handle("/docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("./static/swagger-ui"))))
	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		http.ServeFile(w, r, "./api/v1/openapi.yaml")
	})
}

func (a *API) handleOrders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		a.createOrder(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *API) handleOrderByID(w http.ResponseWriter, r *http.Request) {

	path := strings.TrimPrefix(r.URL.Path, "/orders/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if len(parts) == 1 {
		if r.Method == http.MethodGet {
			a.getOrder(w, r, id)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if len(parts) == 2 && r.Method == http.MethodPost {
		switch parts[1] {
		case "pay":
			a.payOrder(w, r, id)
			return
		case "ship":
			a.shipOrder(w, r, id)
			return
		case "cancel":
			a.cancelOrder(w, r, id)
			return
		default:
			http.NotFound(w, r)
			return
		}
	}

	http.NotFound(w, r)
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

	writeJSON(w, http.StatusCreated, o)
}

func (a *API) getOrder(w http.ResponseWriter, r *http.Request, id int64) {
	o, err := a.orders.GetByID(r.Context(), id)
	if err != nil {
		if err.Error() == "not found" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (a *API) payOrder(w http.ResponseWriter, r *http.Request, id int64) {
	o, err := a.orders.MarkPaid(r.Context(), id)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (a *API) shipOrder(w http.ResponseWriter, r *http.Request, id int64) {
	o, err := a.orders.MarkShipped(r.Context(), id)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func (a *API) cancelOrder(w http.ResponseWriter, r *http.Request, id int64) {
	o, err := a.orders.Cancel(r.Context(), id)
	if err != nil {
		mapDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, o)
}

func mapDomainError(w http.ResponseWriter, err error) {
	msg := err.Error()
	switch {
	case msg == "not found":
		http.Error(w, "not found", http.StatusNotFound)
	case strings.HasPrefix(msg, "invalid transition"):
		http.Error(w, msg, http.StatusConflict)
	default:
		http.Error(w, msg, http.StatusBadRequest)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
