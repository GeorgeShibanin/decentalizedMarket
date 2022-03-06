package handlers

import (
	"decentralizedProject/storage"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

type HTTPHandler struct {
	StorageMu sync.RWMutex
	Storage   storage.Storage
}

type PutRequestData struct {
	/*Weight         storage.Weight         `json:"weight"`
	Size           storage.Size           `json:"size"`
	DeparturePoint storage.DeparturePoint `json:"departurePoint"`
	ReceivePoint   storage.ReceivePoint   `json:"receivePoint"`
	OrderReadyDate storage.ISOTimestamp   `json:"orderReadyDate"`
	*/
	OrderId storage.OrderId `json:"id"`
}

func HandleRoot(rw http.ResponseWriter, r *http.Request) {
	_, err := rw.Write([]byte("Hello from server"))
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	rw.Header().Set("Content-Type", "plain/text")
}

func (h *HTTPHandler) HandleGetOrder(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")

	weight, _ := strconv.ParseFloat(r.URL.Query().Get("weight"), 64)
	size, _ := strconv.ParseFloat(r.URL.Query().Get("volume"), 64)
	departurePoint := r.URL.Query().Get("from")
	receivePoint := r.URL.Query().Get("to")
	orderReadyDate := r.URL.Query().Get("time")

	if weight == 0 && size == 0 && departurePoint == "" && receivePoint == "" && orderReadyDate == "" {
		http.Error(rw, "", http.StatusBadRequest)
		return
	}

	newOrder, err := h.Storage.GetOrder(r.Context(), storage.Weight(weight), storage.Size(size),
		storage.DeparturePoint(departurePoint), storage.ReceivePoint(receivePoint),
		storage.ISOTimestamp(orderReadyDate))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rawResponse, _ := json.Marshal(newOrder)

	_, err = rw.Write(rawResponse)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandler) HandleAcceptDelivery(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	var order PutRequestData
	err := json.NewDecoder(r.Body).Decode(&order)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	status, err := h.Storage.AcceptDelivery(r.Context(), order.OrderId)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	rawResponse, _ := json.Marshal(status)
	_, err = rw.Write(rawResponse)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
}
