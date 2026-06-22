// WebSocket upgrade handler and broadcast helpers. Clients connect with a JWT
// query param; the server resolves their company and subscribes the connection
// to that company's channel. Simulation ticks and notable events broadcast
// updates so the live dashboard refreshes without polling.
package server

import (
	"context"
	"net/http"

	"github.com/coder/websocket"

	"github.com/YASSERRMD/Ventiqra/backend/internal/realtime"
)

// handleWebSocket upgrades the connection and subscribes it to the caller's
// company channel. Authentication is via a `token` query parameter (browsers
// cannot set Authorization headers on WebSocket handshakes).
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	if s.hub == nil || s.tokens == nil || s.companies == nil {
		http.Error(w, "realtime service not configured", http.StatusServiceUnavailable)
		return
	}
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		http.Error(w, "token required", http.StatusUnauthorized)
		return
	}
	claims, err := s.tokens.Parse(tokenStr)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	// Resolve the owner's latest company to subscribe to.
	company, err := s.companies.GetLatestCompanyForOwner(r.Context(), userID)
	if err != nil {
		http.Error(w, "no company found", http.StatusNotFound)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// The frontend is same-origin; allow all origins for dev convenience.
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	// Limit frame sizes to keep memory bounded.
	conn.SetReadLimit(1 << 14) // 16 KiB

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Send an initial hello so the client knows the stream is live.
	s.hub.Publish(realtime.Message{
		Type: "hello", CompanyID: company.ID,
		Payload: map[string]any{
			"company_id": company.ID,
			"name":       company.Name,
			"status":     string(company.Status),
		},
	})

	s.hub.Subscribe(ctx, company.ID, conn)
	_ = conn.CloseNow()
}

// broadcastTick publishes the latest sim state to the company's subscribers
// after a tick completes. Best-effort; failures are ignored.
func (s *Server) broadcastTick(companyID string, day int, cash, revenue, burn int64, customers int) {
	if s.hub == nil {
		return
	}
	s.hub.Publish(realtime.Message{
		Type: "tick", CompanyID: companyID,
		Payload: map[string]any{
			"day":        day,
			"cash":       cash,
			"revenue":    revenue,
			"burn":       burn,
			"customers":  customers,
		},
	})
}

// broadcastEvent publishes a notable event (decision offered, crisis, etc.).
func (s *Server) broadcastEvent(companyID, kind, title string) {
	if s.hub == nil {
		return
	}
	s.hub.Publish(realtime.Message{
		Type: kind, CompanyID: companyID,
		Payload: map[string]string{"title": title},
	})
}
