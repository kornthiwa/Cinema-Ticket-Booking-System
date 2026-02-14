# Cinema Ticket Booking System

Real-time cinema seat booking with distributed lock, WebSocket, and message queue.

---

## How to Run

```bash
docker compose up --build
```

- **App (browser):** http://localhost
- **API:** http://localhost (proxied via nginx to backend)

---

## Seed Data & Login (รันครั้งแรก)

เมื่อรันครั้งแรก ระบบจะ seed ข้อมูลอัตโนมัติ: รอบฉาย 3 เรื่อง + User 2 คน

### ID ที่ใช้ Login

| บทบาท | Email / User ID ที่ใส่ตอน Login |
|--------|----------------------------------|
| **User** (จองที่นั่ง) | `user@cinema.local` |
| **Admin** (จัดการระบบ) | `admin@cinema.local` |

**วิธีใช้:**
- **จองที่นั่ง:** ใส่ `user@cinema.local` → กด **Login** → เข้าหน้า Screenings → เลือกรอบ → เลือกที่นั่ง → Confirm payment
- **เข้า Admin:** ใส่ `admin@cinema.local` → กด **Admin login** → ดู Bookings / Audit logs / สร้างรอบฉาย

Display name ใส่อะไรก็ได้ (เช่น User, Admin)

---

## System Overview

- **Backend:** Go (Gin), MongoDB, Redis (lock + Pub/Sub), WebSocket
- **Frontend:** Vue 3, Vite, nginx (SPA + proxy)
- **Booking:** Lock ที่นั่ง 5 นาที (Redis), ยืนยันแล้วเป็น BOOKED; ไม่จ่ายในเวลา → TIMEOUT แล้วปล่อย lock

---

## Tech Stack

| Layer     | Technology        |
|----------|--------------------|
| Backend  | Go 1.21, Gin      |
| Frontend | Vue 3, Vue Router, Vite |
| Database | MongoDB 7         |
| Lock     | Redis 7 (distributed lock only) |
| Realtime | WebSocket         |
| MQ       | Redis Pub-Sub     |
| Auth     | JWT (mock)        |
| Deploy   | Docker Compose    |

---

## Redis Lock Strategy

- Key: `seat_lock:{screeningID}:{row}:{col}`, value: UUID, TTL: 5 min
- Acquire: `SET key lock_id NX EX 300`
- Release: Lua script — delete only if value matches lock_id
- Prevents double booking when many users select the same seat

---

## Message Queue (Redis Pub-Sub)

- Channel: `booking_events`
- Events: `BOOKING_SUCCESS`, `SEAT_RELEASED`
- Subscriber: audit log to MongoDB + mock notification (log)

---

## Assumptions

- Mock auth (user_id/email + JWT). Production: Google OAuth / Firebase.
- No real payment; "Confirm payment" only marks booking CONFIRMED.
- Admin created via POST /admin/login or seed user `admin@cinema.local`.
