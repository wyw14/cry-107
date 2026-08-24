package api

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	system *System
}

func NewHandler(system *System) *Handler {
	return &Handler{system: system}
}

func (h *Handler) health(writer http.ResponseWriter, _ *http.Request) {
	if err := h.system.ValidateReadyState(); err != nil {
		writeError(writer, http.StatusServiceUnavailable, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"status": "ok", "service": "kilnguard"})
}

func (h *Handler) kilnState(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, h.system.State())
}

func (h *Handler) burnerState(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"burner": h.system.Burner.State(), "damper": h.system.Damper.Snapshot(), "draft": h.system.DraftState()})
}

func (h *Handler) coolerState(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"cooler": h.system.Cooler.Snapshot(), "budget": h.system.FanBudget.Snapshot(), "waste_heat": h.system.WasteHeat.Snapshot(), "recovery": h.system.Recovery.Snapshot()})
}

func (h *Handler) incidents(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{"incidents": h.system.Interlocks.List()})
}

func (h *Handler) requestLoad(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		LoadPercent float64 `json:"load_percent"`
	}
	if err := decodeJSON(request, &body); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	result, err := h.system.RequestLoad(body.LoadPercent)
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusAccepted, result)
}

func (h *Handler) confirmAir(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		ActualAir float64 `json:"actual_air"`
	}
	if err := decodeJSON(request, &body); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	state, err := h.system.ConfirmAir(body.ActualAir)
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, state)
}

func (h *Handler) coolingAir(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		OutletTemp   float64 `json:"outlet_temp"`
		RequestedAir float64 `json:"requested_air"`
	}
	if err := decodeJSON(request, &body); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	state, err := h.system.AllocateCooling(body.OutletTemp, body.RequestedAir)
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, state)
}

func (h *Handler) wasteHeat(writer http.ResponseWriter, request *http.Request) {
	var body struct {
		RequestedAir float64 `json:"requested_air"`
	}
	if err := decodeJSON(request, &body); err != nil {
		writeError(writer, http.StatusBadRequest, err)
		return
	}
	state, err := h.system.AllocateWasteHeat(body.RequestedAir)
	if err != nil {
		writeError(writer, http.StatusConflict, err)
		return
	}
	writeJSON(writer, http.StatusOK, state)
}

func decodeJSON(request *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(nil, request.Body, 1<<20))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, status int, err error) {
	writeJSON(writer, status, map[string]string{"error": err.Error()})
}
