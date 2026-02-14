package repository

import (
	"context"
	"time"

	"cinema-booking/internal/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepo struct {
	db *mongo.Database
}

func NewMongoRepo(db *mongo.Database) *MongoRepo {
	return &MongoRepo{db: db}
}

func (r *MongoRepo) screeningCol() *mongo.Collection { return r.db.Collection("screenings") }
func (r *MongoRepo) bookingCol() *mongo.Collection   { return r.db.Collection("bookings") }
func (r *MongoRepo) userCol() *mongo.Collection     { return r.db.Collection("users") }
func (r *MongoRepo) auditCol() *mongo.Collection    { return r.db.Collection("audit_logs") }

func (r *MongoRepo) AuditCol() *mongo.Collection { return r.auditCol() }

func (r *MongoRepo) CreateScreening(ctx context.Context, s *model.Screening) error {
	if s.CreatedAt.IsZero() {
		s.CreatedAt = time.Now()
	}
	_, err := r.screeningCol().InsertOne(ctx, s)
	return err
}

func (r *MongoRepo) GetScreening(ctx context.Context, id string) (*model.Screening, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var s model.Screening
	err = r.screeningCol().FindOne(ctx, bson.M{"_id": oid}).Decode(&s)
	if err != nil {
		return nil, err
	}
	s.ID = oid
	return &s, nil
}

func (r *MongoRepo) ListScreenings(ctx context.Context) ([]*model.Screening, error) {
	cur, err := r.screeningCol().Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"screen_at": 1}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*model.Screening
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *MongoRepo) CreateBooking(ctx context.Context, b *model.Booking) error {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now()
	}
	res, err := r.bookingCol().InsertOne(ctx, b)
	if err != nil {
		return err
	}
	b.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *MongoRepo) GetBookingByID(ctx context.Context, id string) (*model.Booking, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var b model.Booking
	if err := r.bookingCol().FindOne(ctx, bson.M{"_id": oid}).Decode(&b); err != nil {
		return nil, err
	}
	b.ID = oid
	return &b, nil
}

func (r *MongoRepo) GetBookingByLock(ctx context.Context, screeningID, lockID string) (*model.Booking, error) {
	var b model.Booking
	err := r.bookingCol().FindOne(ctx, bson.M{"screening_id": screeningID, "lock_id": lockID}).Decode(&b)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *MongoRepo) ConfirmBooking(ctx context.Context, bookingID string) error {
	oid, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = r.bookingCol().UpdateOne(ctx, bson.M{"_id": oid},
		bson.M{"$set": bson.M{"status": "CONFIRMED", "confirmed_at": now}})
	return err
}

func (r *MongoRepo) SetBookingStatus(ctx context.Context, bookingID, status string) error {
	oid, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return err
	}
	_, err = r.bookingCol().UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"status": status}})
	return err
}

// SetBookingStatusIfPending sets status to newStatus only if current status is PENDING. Returns true if updated.
func (r *MongoRepo) SetBookingStatusIfPending(ctx context.Context, bookingID, newStatus string) (bool, error) {
	oid, err := primitive.ObjectIDFromHex(bookingID)
	if err != nil {
		return false, err
	}
	res, err := r.bookingCol().UpdateOne(ctx,
		bson.M{"_id": oid, "status": "PENDING"},
		bson.M{"$set": bson.M{"status": newStatus}})
	if err != nil {
		return false, err
	}
	return res.ModifiedCount == 1, nil
}

func (r *MongoRepo) ListBookings(ctx context.Context, filter bson.M) ([]*model.Booking, error) {
	cur, err := r.bookingCol().Find(ctx, filter, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []*model.Booking
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *MongoRepo) UpsertUser(ctx context.Context, u *model.User) error {
	set := bson.M{"email": u.Email, "name": u.Name, "role": u.Role}
	if u.PasswordHash != "" {
		set["password_hash"] = u.PasswordHash
	}
	_, err := r.userCol().UpdateOne(ctx, bson.M{"_id": u.ID},
		bson.M{"$set": set},
		options.Update().SetUpsert(true))
	return err
}

func (r *MongoRepo) GetUser(ctx context.Context, id string) (*model.User, error) {
	var u model.User
	err := r.userCol().FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *MongoRepo) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.userCol().FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *MongoRepo) InsertAuditLog(ctx context.Context, event string, payload map[string]any) error {
	doc := model.AuditLog{Event: event, Payload: payload, CreatedAt: time.Now()}
	_, err := r.auditCol().InsertOne(ctx, doc)
	return err
}
