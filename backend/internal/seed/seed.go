package seed

import (
	"context"
	"log"
	"time"

	"cinema-booking/internal/model"
	"cinema-booking/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Run inserts seed data when the database has no screenings (first run).
func Run(ctx context.Context, repo *repository.MongoRepo) {
	list, err := repo.ListScreenings(ctx)
	if err != nil {
		log.Printf("seed: list screenings: %v", err)
		return
	}
	if len(list) > 0 {
		return
	}

	log.Println("seed: first run — inserting seed data")

	now := time.Now()
	screenings := []*model.Screening{
		{
			ID:        primitive.NewObjectID(),
			MovieID:   "mv-001",
			MovieName: "The Matrix",
			ScreenAt:  now.Add(24 * time.Hour),
			Rows:      5,
			Cols:      8,
			CreatedAt: now,
		},
		{
			ID:        primitive.NewObjectID(),
			MovieID:   "mv-002",
			MovieName: "Inception",
			ScreenAt:  now.Add(48 * time.Hour),
			Rows:      6,
			Cols:      10,
			CreatedAt: now,
		},
		{
			ID:        primitive.NewObjectID(),
			MovieID:   "mv-003",
			MovieName: "Interstellar",
			ScreenAt:  now.Add(72 * time.Hour),
			Rows:      5,
			Cols:      8,
			CreatedAt: now,
		},
	}

	for _, s := range screenings {
		if err := repo.CreateScreening(ctx, s); err != nil {
			log.Printf("seed: create screening %s: %v", s.MovieName, err)
			continue
		}
		log.Printf("seed: created screening %s (%s)", s.MovieName, s.ID.Hex())
	}

	users := []*model.User{
		{ID: "admin@cinema.local", Email: "admin@cinema.local", Name: "Admin", Role: model.RoleAdmin},
		{ID: "user@cinema.local", Email: "user@cinema.local", Name: "Demo User", Role: model.RoleUser},
	}
	for _, u := range users {
		if err := repo.UpsertUser(ctx, u); err != nil {
			log.Printf("seed: upsert user %s: %v", u.Email, err)
			continue
		}
		log.Printf("seed: created user %s (%s)", u.Email, u.Role)
	}

	log.Println("seed: done")
}
