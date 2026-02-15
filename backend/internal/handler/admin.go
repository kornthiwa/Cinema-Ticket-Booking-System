package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (h *Handler) ListBookingsAdmin(c *gin.Context) {
	filter := bson.M{}
	if userID := c.Query("user_id"); userID != "" {
		filter["user_id"] = userID
	}
	if screeningID := c.Query("screening_id"); screeningID != "" {
		filter["screening_id"] = screeningID
	} else if movieName := strings.TrimSpace(c.Query("movie_name")); movieName != "" {
		screenings, _ := h.Repo.ListScreenings(c.Request.Context())
		var ids []string
		lower := strings.ToLower(movieName)
		for _, s := range screenings {
			if strings.Contains(strings.ToLower(s.MovieName), lower) {
				ids = append(ids, s.ID.Hex())
			}
		}
		if len(ids) > 0 {
			filter["screening_id"] = bson.M{"$in": ids}
		} else {
			filter["screening_id"] = "none"
		}
	} else if movieID := c.Query("movie_id"); movieID != "" {
		screenings, _ := h.Repo.ListScreenings(c.Request.Context())
		var ids []string
		for _, s := range screenings {
			if s.MovieID == movieID {
				ids = append(ids, s.ID.Hex())
			}
		}
		if len(ids) > 0 {
			filter["screening_id"] = bson.M{"$in": ids}
		} else {
			filter["screening_id"] = "none"
		}
	}
	list, err := h.Repo.ListBookings(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Enrich with screening info for display
	type row struct {
		Booking   interface{} `json:"booking"`
		MovieName string     `json:"movie_name,omitempty"`
	}
	out := make([]row, len(list))
	for i, b := range list {
		out[i] = row{Booking: b}
		if s, _ := h.Repo.GetScreening(c.Request.Context(), b.ScreeningID); s != nil {
			out[i].MovieName = s.MovieName
		}
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) ListAuditLogs(c *gin.Context) {
	// Simple list from MongoDB; could add pagination
	col := h.Repo.AuditCol()
	cur, err := col.Find(c.Request.Context(), bson.M{}, options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(200))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(c.Request.Context())
	var logs []map[string]interface{}
	if err := cur.All(c.Request.Context(), &logs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, logs)
}
