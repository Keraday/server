package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"server/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionHandler struct {
	DB  *pgxpool.Pool
	Log *slog.Logger
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.Log.Warn("failed to read r.Body", slog.Any("error", err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var sub models.Subscription

	err = json.Unmarshal(body, &sub)
	if err != nil {
		h.Log.Warn("failed to unmarshal r.Body", slog.Any("error", err))
		http.Error(w, fmt.Sprintf("invalid JSON: %v", err), http.StatusBadRequest)
		return
	}
	if sub.ServiceName == "" {
		http.Error(w, "service_name is required", http.StatusBadRequest)
		return
	}
	if sub.Price < 0 {
		http.Error(w, "price must be non-negative", http.StatusBadRequest)
		return
	}
	if sub.UserID == uuid.Nil {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}
	if sub.StartDate == "" {
		http.Error(w, "start_date is required", http.StatusBadRequest)
		return
	}

	startDateTime, err := parseStringToTime(sub.StartDate)
	if err != nil {
		h.Log.Warn("invalid start date format", slog.Any("error", err))
		http.Error(w, "invalid start date format, expected MM-YYYY", http.StatusBadRequest)
		return
	}
	var endDateTime *time.Time
	if sub.EndDate != nil && *sub.EndDate != "" {
		t, err := parseStringToTime(*sub.EndDate)
		if err != nil {
			h.Log.Warn("invalid end date format", slog.Any("error", err))
			http.Error(w, "invalid end date format, expected MM-YYYY", http.StatusBadRequest)
			return
		}
		endDateTime = &t
	}
	sub.ID = uuid.New()

	h.Log.Debug("create", "unmarshaled body", sub)

	query := `
    INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
    SELECT  $1::uuid, $2::text, $3::numeric, $4::uuid, $5::date, $6::date
    WHERE NOT EXISTS (
        SELECT 1
        FROM subscriptions
        WHERE user_id = $4
          AND service_name = $2
          AND start_date < COALESCE($6, 'infinity'::date)
          AND COALESCE(end_date, 'infinity'::date) > $5
    )
    RETURNING 1`
	var dummy int
	err = h.DB.QueryRow(r.Context(), query,
		sub.ID, sub.ServiceName, sub.Price, sub.UserID, startDateTime, endDateTime,
	).Scan(&dummy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.Log.Warn("subscription period overlaps with existing one",
				"user_id", sub.UserID,
				"service", sub.ServiceName,
				"start_date", sub.StartDate)
			http.Error(w, "subscription period overlaps with existing one", http.StatusConflict)
			return
		}
		h.Log.Error("failed to insert DB", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	h.Log.Info("subscription created", "id", sub.ID, "user_id", sub.UserID, "service", sub.ServiceName)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(sub); err != nil {
		h.Log.Error("failed to write response", slog.Any("error", err))
	}
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.Log.Warn("invalid id format", slog.Any("id", idStr))
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}
	h.Log.Debug("DELETE", "received id", id)

	res, err := h.DB.Exec(r.Context(), "DELETE FROM subscriptions WHERE id=$1", id)
	if err != nil {
		h.Log.Error("failed to delete subscription", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if res.RowsAffected() == 0 {
		http.Error(w, "subscription not found", http.StatusNotFound)
		return
	}
	h.Log.Info("subscription deleted", "id", id)

	w.WriteHeader(http.StatusNoContent)
}

func (h *SubscriptionHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		h.Log.Warn("no {id} for GetByID")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.Log.Warn("invalid id format", slog.Any("error", err))
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}
	var sub models.Subscription
	var startDatetime time.Time
	var endDateTime *time.Time

	err = h.DB.QueryRow(r.Context(), "SELECT id, service_name, price, user_id, start_date, end_date FROM subscriptions WHERE id=$1",
		id).Scan(&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID, &startDatetime, &endDateTime)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "subscription not found", http.StatusNotFound)
			return
		}
		h.Log.Error("failed to fetch subscription", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	sub.StartDate = startDatetime.Format("01-2006")
	if endDateTime == nil {
		sub.EndDate = nil
	} else {
		s := endDateTime.Format("01-2006")
		sub.EndDate = &s
	}
	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(sub)
	if err != nil {
		h.Log.Error("failed to write response", slog.Any("error", err))
	}
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		h.Log.Warn("no {id} for Update")
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.Log.Warn("invalid id format", slog.Any("error", err))
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.Log.Warn("failed to read r.Body", slog.Any("error", err))
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var input struct {
		Price   int     `json:"price"`
		EndDate *string `json:"end_date,omitempty"`
	}
	err = json.Unmarshal(body, &input)
	if err != nil {
		h.Log.Warn("failed to unmarshal r.Body", slog.Any("error", err))
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Price < 0 {
		http.Error(w, "price must be non-negative", http.StatusBadRequest)
		return
	}

	var pgEndDate *time.Time
	if input.EndDate != nil && *input.EndDate != "" {
		t, err := parseStringToTime(*input.EndDate)
		if err != nil {
			h.Log.Warn("invalid end date format", slog.Any("error", err), "value", *input.EndDate)
			http.Error(w, "invalid end date format, expected MM-YYYY", http.StatusBadRequest)
			return
		}
		pgEndDate = &t
	}

	query := `
	UPDATE subscriptions 
	SET price=$1, end_date=$2 
	WHERE id=$3 
	RETURNING id, service_name, price, user_id, start_date, end_date`
	var res models.Subscription
	var startDateTime time.Time
	var endDateTime *time.Time

	err = h.DB.QueryRow(r.Context(), query, input.Price, pgEndDate, id).
		Scan(&res.ID, &res.ServiceName, &res.Price, &res.UserID, &startDateTime, &endDateTime)
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "subscription not found", http.StatusNotFound)
			return
		}
		h.Log.Error("failed to update DB", slog.Any("error", err))
		http.Error(w, "failed to update DB", http.StatusInternalServerError)
		return
	}
	res.StartDate = startDateTime.Format("01-2006")

	if endDateTime == nil {
		res.EndDate = nil
	} else {
		s := endDateTime.Format("01-2006")
		res.EndDate = &s
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		h.Log.Error("failed to write response", slog.Any("error", err))
	}
}

func (h *SubscriptionHandler) GetTotal(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.Log.Warn("invalid user_id format", "user_id", userIDStr)
		http.Error(w, "invalid user_id format", http.StatusBadRequest)
		return
	}

	serviceName := r.URL.Query().Get("service_name")

	fromStr := r.URL.Query().Get("from")
	if fromStr == "" {
		http.Error(w, "from date is required", http.StatusBadRequest)
		return
	}

	fromTime, err := parseStringToTime(fromStr)
	if err != nil {
		h.Log.Warn("invalid from date format", "value", fromStr)
		http.Error(w, "invalid from date format, expected MM-YYYY", http.StatusBadRequest)
		return
	}

	toTime := time.Now()
	toStr := r.URL.Query().Get("to")
	if toStr != "" {
		toTime, err = parseStringToTime(toStr)
		if err != nil {
			h.Log.Warn("invalid to date format", "value", toStr)
			http.Error(w, "invalid to date format, expected MM-YYYY", http.StatusBadRequest)
			return
		}
	}
	if !fromTime.Before(toTime) {
		h.Log.Warn("invalid date range: from must be strictly before to",
			"from", fromStr,
			"to", toStr)
		http.Error(w, "from date must be strictly before to date", http.StatusBadRequest)
		return
	}
	query := `
        SELECT COALESCE(SUM(
            price * (
                (EXTRACT(YEAR FROM AGE(
                    LEAST(COALESCE(end_date, $3), $3),
                    GREATEST(start_date, $2)
                )) * 12 +
                EXTRACT(MONTH FROM AGE(
                    LEAST(COALESCE(end_date, $3), $3),
                    GREATEST(start_date, $2)
                )))::INTEGER + 1
            )
        ), 0)::INTEGER AS total
        FROM subscriptions
        WHERE user_id = $1
          AND start_date < $3
          AND COALESCE(end_date, $3) > $2
    `

	args := []interface{}{userID, fromTime, toTime}

	if serviceName != "" {
		query = `
            SELECT COALESCE(SUM(
                price * (
                    (EXTRACT(YEAR FROM AGE(
                        LEAST(COALESCE(end_date, $4), $4),
                        GREATEST(start_date, $3)
                    )) * 12 +
                    EXTRACT(MONTH FROM AGE(
                        LEAST(COALESCE(end_date, $4), $4),
                        GREATEST(start_date, $3)
                    )))::INTEGER + 1
                )
            ), 0)::INTEGER AS total
            FROM subscriptions
            WHERE user_id = $1
              AND service_name = $2
              AND start_date < $4
              AND COALESCE(end_date, $4) > $3
        `
		args = []interface{}{userID, serviceName, fromTime, toTime}
	}

	var total int
	err = h.DB.QueryRow(r.Context(), query, args...).Scan(&total)
	if err != nil {
		h.Log.Error("failed to calculate total", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]int{"total": total})
	if err != nil {
		h.Log.Error("failed to write response", slog.Any("error", err))
	}
}

func parseStringToTime(strTime string) (time.Time, error) {
	return time.Parse("01-2006", strTime)
}
