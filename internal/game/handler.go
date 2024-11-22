package game

import (
	"encoding/json"
	"net/http"
	
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"

	"big-spella-go/internal/auth"
)

type Handler struct {
	service  GameService
	upgrader websocket.Upgrader
}

func NewHandler(service GameService) *Handler {
	return &Handler{
		service: service,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // TODO: Add proper origin check
			},
		},
	}
}

type CreateGameRequest struct {
	Type     GameType     `json:"type"`
	Settings GameSettings `json:"settings"`
}

func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	game, err := h.service.CreateGame(r.Context(), userID, req.Type, req.Settings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(game)
}

func (h *Handler) JoinGame(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	gameID := ps.ByName("gameID")
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	game, err := h.service.JoinGame(r.Context(), gameID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(game)
}

func (h *Handler) StartGame(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	gameID := ps.ByName("gameID")
	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	game, err := h.service.StartGame(r.Context(), gameID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(game)
}

type MakeAttemptRequest struct {
	Type      AttemptType `json:"type"`
	Text      *string     `json:"text,omitempty"`
	VoiceData []byte      `json:"voice_data,omitempty"`
}

func (h *Handler) MakeAttempt(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	gameID := ps.ByName("gameID")
	if gameID == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req MakeAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var attempt *SpellingAttempt
	switch req.Type {
	case AttemptTypeText:
		if req.Text == nil {
			http.Error(w, "Text is required for text attempt", http.StatusBadRequest)
			return
		}
		attempt = &SpellingAttempt{
			Type: AttemptTypeText,
			Text: *req.Text,
		}
	case AttemptTypeVoice:
		if len(req.VoiceData) == 0 {
			http.Error(w, "Voice data is required for voice attempt", http.StatusBadRequest)
			return
		}
		attempt = &SpellingAttempt{
			Type:      AttemptTypeVoice,
			VoiceData: req.VoiceData,
		}
	default:
		http.Error(w, "Invalid attempt type", http.StatusBadRequest)
		return
	}

	if err := h.service.MakeAttempt(r.Context(), gameID, userID, attempt); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetGame(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	gameID := ps.ByName("gameID")
	if gameID == "" {
		http.Error(w, "Game ID is required", http.StatusBadRequest)
		return
	}

	userID := auth.GetUserIDFromContext(r.Context())
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	game, err := h.service.GetGame(r.Context(), gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(game)
}

func (h *Handler) SubscribeToEvents(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	for event := range h.service.Events() {
		if err := conn.WriteJSON(event); err != nil {
			break
		}
	}
}

func (h *Handler) Routes() *httprouter.Router {
	router := httprouter.New()

	router.POST("/games", h.CreateGame)
	router.POST("/games/:gameID/join", h.JoinGame)
	router.POST("/games/:gameID/start", h.StartGame)
	router.POST("/games/:gameID/attempt", h.MakeAttempt)
	router.GET("/games/:gameID", h.GetGame)
	router.GET("/games/:gameID/events", h.SubscribeToEvents)

	return router
}
