# Cinema Ticket Booking System

Take-Home Assignment: ระบบจองตั๋วหนังออนไลน์ — ออกแบบให้รองรับการแย่งกันซื้อ (Real-time seat map, Redis Distributed Lock, WebSocket, Message Queue)

---

## 1. System Architecture Diagram

```
                    ┌─────────────────────────────────────────────────────────┐
                    │                     Docker Compose                       │
  Browser           │                                                         │
     │              │  ┌──────────┐     ┌─────────┐     ┌─────────┐           │
     │  :80         │  │  nginx   │────▶│ Backend │────▶│ MongoDB │           │
     └──────────────┼─▶│ (Vue SPA │     │  (Gin)  │     │  (DB)   │           │
                    │  │ + proxy) │     └────┬────┘     └─────────┘           │
                    │  └──────────┘          │                                 │
                    │                        │  Lock / Pub-Sub                  │
                    │                        ▼                                 │
                    │                   ┌─────────┐                            │
                    │                   │  Redis  │  ◀── Worker (lock expiry)  │
                    │                   └─────────┘                            │
                    │                        │                                 │
                    │              WebSocket │  Real-time seat updates         │
                    └───────────────────────┼─────────────────────────────────┘
                                            │
                              Frontend (Vue 3) ◀── WS + REST
```

- **User:** Login → เลือกรอบฉาย → เห็นผังที่นั่ง Real-time (WebSocket) → เลือกที่นั่ง (Lock) → ชำระเงิน → BOOKED
- **Admin:** Login → Dashboard (Bookings + Filter) / Audit Logs / สร้างรอบฉาย
- **Backend:** REST API, WebSocket Hub, Redis Lock, Redis Pub-Sub subscriber (Audit + Mock Notification), Lock Expiry Worker

---

## 2. Tech Stack Overview

| Layer        | Technology                          |
|-------------|--------------------------------------|
| Backend     | Go (Gin)                             |
| Frontend    | Vue 3, Vue Router, Vite              |
| Database    | MongoDB                              |
| Cache/Lock  | Redis (Distributed Lock เท่านั้น ไม่ใช้เป็น cache) |
| Realtime    | WebSocket                            |
| Message Queue | Redis Pub-Sub                      |
| Auth        | JWT (mock); Production ใช้ Google OAuth / Firebase ได้ |
| Deployment  | Docker + docker-compose.yml          |

---

## 3. Booking Flow (อธิบายทีละ Step)

| Step | ผู้ใช้ | ระบบ (Backend) |
|------|--------|-----------------|
| 1 | เลือกที่นั่งบนผัง | รับ request lock ที่นั่ง |
| 2 | — | **Redis Distributed Lock:** `SET key NX EX 300` (5 นาที), key = `seat_lock:{screeningID}:{row}:{col}`, value = lock_id (UUID) |
| 3 | — | ถ้า lock สำเร็จ → สถานะที่นั่งเป็น **LOCKED**, broadcast ผ่าน WebSocket ให้ทุก client เห็นแบบ Real-time; ถ้า lock ไม่ได้ (มีคนถืออยู่) → คืน error |
| 4 | เห็นที่นั่งเป็น LOCKED (และผู้ใช้คนอื่นเห็นเช่นกัน) | — |
| 5 | กด "Confirm payment" ภายใน 5 นาที | ตรวจสอบ lock_id ตรงกับที่ถืออยู่ → อัปเดต booking เป็น **CONFIRMED** (ที่นั่งเป็น **BOOKED**) → ปล่อย lock (Lua script ลบ key ถ้า value ตรง) → Publish `BOOKING_SUCCESS` ไป MQ |
| 6a | ไม่ชำระภายใน 5 นาที | **Worker (Lock Expiry):** เจอ PENDING หมดอายุ → ปล่อย Redis lock → อัปเดต booking เป็น TIMEOUT → Publish `SEAT_RELEASED` → Broadcast ผังใหม่ผ่าน WebSocket |
| 6b | — | MQ Subscriber: รับ event → บันทึก **Audit Log** (MongoDB) + Mock Notification (log) |

ผลลัพธ์: **ไม่มี Double Booking** — ที่นั่งเดียวกันจะ lock ได้เพียงหนึ่ง client ในช่วง 5 นาที

---

## 4. Redis Lock Strategy

- **Key:** `seat_lock:{screeningID}:{row}:{col}`
- **Value:** UUID (lock_id) ของ client ที่ถือ lock
- **TTL:** 300 วินาที (5 นาที) — กำหนดใน env `LOCK_TTL_SECONDS`
- **Acquire:** `SET key lock_id NX EX 300` — สร้าง key ได้เฉพาะเมื่อยังไม่มี (NX) จึงกันการแย่งกัน lock ที่นั่งเดียวกัน
- **Release:** Lua script ลบ key **เฉพาะเมื่อ value ตรงกับ lock_id** — ป้องกันการปล่อย lock ของคนอื่น (เช่น หลัง timeout แล้วมีคนใหม่มา lock)
- **เหตุผลออกแบบ:** ใช้ Redis แบบ single-key lock + TTL เพื่อให้หลาย instance ของ backend รองรับ concurrency ได้โดยไม่มี double booking และไม่ต้องพึ่ง in-memory lock แค่ process เดียว

---

## 5. Message Queue ใช้ทำอะไร

- **Broker:** Redis Pub-Sub (เลือกตามโจทย์ 1 ใน Kafka / RabbitMQ / Redis Pub-Sub)
- **Channel:** `booking_events`
- **Events ที่ publish:** `BOOKING_SUCCESS`, `SEAT_RELEASED` (จาก backend ตอน confirm booking และจาก worker ตอน timeout)
- **Use case จริงที่ใช้ MQ:**
  1. **Audit Log:** Subscriber รับ event แล้วบันทึกลง MongoDB (`audit_logs`) — เช่น Booking Success, Booking Timeout, Seat Released
  2. **Mock Notification:** เมื่อได้ event `BOOKING_SUCCESS` → log เป็น mock notification (ใน production ต่อ Email / Line ได้)
- ไม่ใช้ MQ แค่มีไว้เฉย ๆ — ทุก event จาก booking/lock flow ส่งไป MQ และมี subscriber process จริง

---

## 6. วิธีรันระบบ

ต้องรันได้ด้วยคำสั่งเดียว:

```bash
docker compose up --build
```

- **แอป (Frontend):** http://localhost  
- **API:** ใช้ผ่าน http://localhost (nginx proxy ไปที่ backend)

### Seed Data & Login (รันครั้งแรก)

เมื่อรันครั้งแรก ระบบจะ seed ข้อมูลอัตโนมัติ: รอบฉาย 3 เรื่อง + User 2 คน

| บทบาท | Email / User ID ที่ใส่ตอน Login |
|--------|----------------------------------|
| **User** (จองที่นั่ง) | `user@cinema.local` |
| **Admin** (จัดการระบบ) | `admin@cinema.local` |

**วิธีใช้:**
- **จองที่นั่ง:** ใส่ `user@cinema.local` → กด **Login** → เข้าหน้า Screenings → เลือกรอบ → เลือกที่นั่ง → Confirm payment
- **เข้า Admin:** ใส่ `admin@cinema.local` → กด **Admin login** → ดู Bookings (มี filter) / Audit logs / สร้างรอบฉาย

Display name ใส่อะไรก็ได้ (เช่น User, Admin)

---

## 7. Assumptions & Trade-offs

**Assumptions**

- **Auth:** ใช้ mock (user_id/email + JWT). โจทย์กำหนด Google OAuth หรือ Firebase — ระบบออกแบบให้ต่อ Firebase ได้ (env `FIREBASE_PROJECT_ID`); ถ้าไม่ตั้ง จะใช้ mock login กับ seed user
- **Payment:** ไม่มี gateway จริง — "Confirm payment" คือการกดยืนยันในระบบ แล้วเปลี่ยนสถานะเป็น CONFIRMED
- **Admin:** สร้างผ่าน seed (`admin@cinema.local`) หรือ flow Admin login; Admin API ตรวจ role ไม่ให้ User เรียก

**Trade-offs**

- ใช้ **Redis Pub-Sub** แทน Kafka/RabbitMQ เพื่อให้รันด้วย `docker compose` ง่าย ไม่ต้องเพิ่ม service; ถ้า scale ใหญ่ขึ้นอาจเปลี่ยนเป็น Kafka สำหรับ audit/event
- **Lock TTL 5 นาที** เป็นค่าคงที่ในโจทย์; อ่านจาก env ได้ (`LOCK_TTL_SECONDS`) เพื่อปรับใน production
- **Audit events:** บันทึก BOOKING_SUCCESS, BOOKING_TIMEOUT, SEAT_RELEASED, LOCK_FAIL (System Error) ลง MongoDB ผ่าน MQ subscriber

---

## สรุปความสอดคล้องกับโจทย์

| ข้อกำหนด | การทำในโปรเจกต์ |
|----------|-------------------|
| Tech Stack ตามที่กำหนด | Go (Gin), Vue 3, MongoDB, Redis (lock only), WebSocket, Redis Pub-Sub, Docker Compose |
| รันด้วยคำสั่งเดียว | `docker compose up --build` |
| User: Auth, Seat Map Real-time, Booking + Lock 5 นาที | Mock/seed login, WebSocket ผังที่นั่ง, Redis lock 5 นาที, Confirm = BOOKED |
| Admin: Dashboard + Filter, Audit Logs | หน้า Admin: Bookings (filter ได้), Audit logs, สร้างรอบฉาย |
| MQ use case จริง | Publish BOOKING_SUCCESS / SEAT_RELEASED → Audit log + Mock notification |
| Concurrency / No double booking | Redis NX + TTL + release by lock_id |
| Role USER / ADMIN, Admin API แยก | JWT + role; route /admin/* ตรวจสิทธิ์ |
| Config ไม่ hardcode | ใช้ env (MONGODB_URI, REDIS_ADDR, JWT_SECRET, LOCK_TTL_SECONDS, …) |
