# ระบบจองตั๋วหนังออนไลน์ (Cinema Ticket Booking)

โปรเจกต์ Take-Home: ระบบจองตั๋วหนังที่รองรับการแย่งกันซื้อ — มีผังที่นั่งอัปเดตแบบ real-time, ล็อกที่นั่งด้วย Redis, แจ้งเตือนบนหน้าเว็บผ่าน Message Queue

---

## 1. System Architecture Diagram

```
                    ┌─────────────┐
                    │   Browser   │
                    │ (Vue 3 SPA) │
                    └──────┬──────┘
                           │ HTTP / WebSocket
                           ▼
                    ┌─────────────┐
                    │   nginx     │  :80 — เสิร์ฟ static + proxy /api, /auth, /admin, WS
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         ▼                 ▼                 ▼
  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
  │   Backend   │   │   MongoDB   │   │    Redis    │
  │  (Go / Gin) │──▶│  (ข้อมูล    │   │  Lock + MQ  │
  │             │   │   ถาวร)     │   │  (Pub/Sub)  │
  └──────┬──────┘   └─────────────┘   └──────┬──────┘
         │                                    │
         │  Acquire/Release lock, Publish     │ Subscribe
         │  booking_events                   │
         ▼                                    ▼
  ┌─────────────┐                      ┌─────────────┐
  │ WS Hub      │◀─────────────────────│  MQ Sub     │
  │ Broadcast   │  บันทึก Audit Log +   │ (ใน process │
  │ ผังที่นั่ง   │  ส่ง NOTIFICATION     │  backend)   │
  │ + แจ้งเตือน  │  ไป WebSocket         └──────┬──────┘
  └─────────────┘                               │
         ▲                            Worker: ปล่อยล็อกเมื่อหมดเวลา
         │                                     Publish SEAT_RELEASED
         └─────────────────────────────────────┘
```

- **nginx:** รับที่พอร์ต 80, เสิร์ฟ Vue และ proxy ไป backend
- **Backend (Gin):** API + WebSocket, ล็อก/ปล่อยที่นั่งผ่าน Redis, บันทึก MongoDB, Publish event
- **Redis:** Distributed Lock (ที่นั่ง) + Pub/Sub ช่อง `booking_events`
- **Worker (Lock Expiry):** ตรวจจองค้างเกิน TTL → ปล่อยล็อก, อัปเดตสถานะ, Publish + Broadcast

---

## 2. Tech Stack Overview

| ส่วน          | เทคโนโลยี               | หมายเหตุ                                 |
| ------------- | ----------------------- | ---------------------------------------- |
| **Backend**   | Go, Gin                 | REST API + WebSocket                     |
| **Frontend**  | Vue 3, Vue Router, Vite | SPA, หน้า Screenings / SeatMap / Admin   |
| **ฐานข้อมูล** | MongoDB                 | รอบฉาย, การจอง, ผู้ใช้, audit_logs       |
| **ล็อก / MQ** | Redis                   | Distributed Lock (SetNX + TTL) + Pub/Sub |
| **Real-time** | WebSocket               | อัปเดตผังที่นั่งและแจ้งเตือนทันที        |
| **Auth**      | JWT                     | ใช้ใน header / query สำหรับ API และ WS   |
| **รันระบบ**   | Docker Compose          | backend, frontend (nginx), mongo, redis  |

---

## 3. Booking Flow (อธิบายทีละ Step)

| Step  | ผู้ใช้/ระบบ   | รายละเอียด                                                                                                                                                                               |
| ----- | ------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **1** | ผู้ใช้        | ล็อกอิน → เลือกรอบฉาย (Screenings) → เข้าหน้าแผนที่นั่ง (SeatMap)                                                                                                                        |
| **2** | ผู้ใช้        | คลิกที่นั่งที่ว่าง (Available)                                                                                                                                                           |
| **3** | Backend       | รับ `POST /api/screenings/:id/lock` (row, col) → ขอล็อกที่ Redis ด้วย key `seat_lock:{screeningID}:{row}:{col}` (SetNX, TTL 5 นาที)                                                      |
| **4** | Backend       | ถ้าล็อกได้: สร้าง booking สถานะ PENDING ใน MongoDB, Broadcast สถานะ LOCKED ผ่าน WebSocket → ทุก client เห็นที่นั่งเป็นสีเหลือง                                                           |
| **5** | Backend       | ถ้าล็อกไม่ได้ (มีคนถืออยู่): คืน 409 Conflict                                                                                                                                            |
| **6** | ผู้ใช้        | เห็นบล็อก "ที่นั่งที่คุณล็อก" — เลือกรายการที่ต้องการแล้วกด **Confirm payment (mock)**                                                                                                   |
| **7** | Backend       | รับ `POST /api/bookings/confirm` (booking_id) → ตรวจว่าเป็นคนที่ถือล็อก → อัปเดต booking เป็น CONFIRMED, ปล่อยล็อก Redis (Lua script ตรวจ ownership), Publish `BOOKING_SUCCESS` ไป Redis |
| **8** | MQ Subscriber | รับ event → บันทึก Audit Log ลง MongoDB + ส่ง NOTIFICATION ผ่าน WebSocket → หน้าเว็บแสดง "การจองสำเร็จ"                                                                                  |
| **9** | (ทางเลือก)    | ถ้าไม่กดชำระภายใน 5 นาที: Worker ตรวจเจอ booking PENDING หมดอายุ → ปล่อยล็อก, อัปเดตเป็น TIMEOUT, Publish `SEAT_RELEASED`, Broadcast ผังใหม่                                             |

**ผลลัพธ์:** ที่นั่งเดียวกันถูกยึดได้เพียงหนึ่งคนในหนึ่งช่วงเวลา → ไม่เกิด double booking

---

## 4. Redis Lock Strategy

- **Key:** `seat_lock:{screeningID}:{row}:{col}` — หนึ่ง key ต่อหนึ่งที่นั่งต่อหนึ่งรอบ
- **Value:** UUID (lockID) ของผู้ถือล็อก — ใช้ตรวจ ownership ตอนปล่อย
- **TTL:** 5 นาที (กำหนดจาก env `LOCK_TTL_SECONDS=300`) — หมดอายุแล้ว Redis ลบ key อัตโนมัติ
- **Acquire:** `SET key lockID NX EX TTL` — สร้างได้เฉพาะเมื่อ key ยังไม่มี → คนแรกที่ยึดได้เท่านั้น
- **Release:** สคริปต์ Lua — ลบ key **เฉพาะเมื่อ value ตรงกับ lockID ที่ส่งมา** → ไม่ปล่อยล็อกของคนอื่น
- **GetLockID:** อ่าน value ปัจจุบัน — ใช้ตอนแสดงรายการ "ที่นั่งถูกล็อก" และตรวจว่า lock ยังเป็นของ booking นั้นหรือไม่

ทำไมใช้ Redis ไม่ใช่แค่ตัวแปรใน process? เพราะ backend อาจรันหลาย instance — ต้องมีที่กลางให้ทุก instance ขอล็อกที่เดียวกัน จึงใช้ Redis เป็น distributed lock

---

## 5. Message Queue ใช้ทำอะไร

ระบบใช้ **Redis Pub/Sub** เป็นช่องส่งเหตุการณ์ (Message Queue)

- **Channel:** `booking_events`
- **เหตุการณ์:** `BOOKING_SUCCESS` (จองสำเร็จ), `SEAT_RELEASED` (ปล่อยที่นั่ง เช่น หมดเวลา)
- **ผู้ Publish:** Backend (ตอน Confirm payment) และ Worker (ตอนปล่อยล็อกเพราะหมดเวลา)

**เมื่อมี event ใน MQ แล้วเกิดอะไรต่อ:**

1. **Audit Log:** Subscriber (รันใน process เดียวกับ backend) รับ event → บันทึกลง MongoDB (event, payload, เวลา) — ใช้ดูใน Admin
2. **แจ้งเตือน Frontend:** Subscriber ส่งข้อความประเภท NOTIFICATION ไปยัง WebSocket Hub → client ที่เปิดหน้ารอบฉายนั้นอยู่จะเห็นข้อความ "การจองสำเร็จ" / "มีการปล่อยที่นั่ง"

ข้อดี: แยกส่วนงาน — API ไม่ต้องรอให้บันทึก log หรือส่งแจ้งเตือนเสร็จก่อนตอบ client

---

## 6. วิธีรันระบบ

```bash
docker compose up --build
```

- เปิดเว็บ: **http://localhost**
- API และ WebSocket เรียกผ่าน http://localhost (nginx proxy ไป backend)

### ข้อมูลทดสอบ (seed ครั้งแรก)

| บทบาท             | อีเมล                | รหัสผ่าน |
| ----------------- | -------------------- | -------- |
| User (จองที่นั่ง) | `user@cinema.local`  | `123456` |
| Admin             | `admin@cinema.local` | `123456` |

**ลองใช้:** ล็อกอินด้วย user → Screenings → เลือกรอบ → คลิกที่นั่ง → เลือกรายการในบล็อก "ที่นั่งที่คุณล็อก" → กด **Confirm payment (mock)**  
Admin: ล็อกอินแอดมิน → ดู Bookings / Audit logs / สร้างรอบฉาย

---

## 7. Assumptions & Trade-offs

| หัวข้อ          | สมมติฐาน / Trade-off                                                                                                                   |
| --------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| **การชำระเงิน** | ใช้ mock — ไม่มี gateway จริง; เน้น flow ล็อก → ยืนยันจอง                                                                              |
| **Auth**        | JWT แบบง่าย; production ควรใช้ OAuth/Firebase หรือ session ที่ปลอดภัยขึ้น                                                              |
| **User ID**     | ใช้เป็น email หรือ ID จาก auth — เก็บใน booking เพื่อแสดง "ผู้ล็อก/ผู้จอง"                                                             |
| **Real-time**   | WebSocket ต่อห้องต่อรอบฉาย (`screening:{id}`) — ถ้า server restart client ต้อง reconnect เอง                                           |
| **Lock TTL**    | 5 นาทีคงที่ (ปรับได้) — ไม่มีปุ่ม "ขยายเวลา" ใน UI ตัวอย่าง                                                                            |
| **MQ**          | Redis Pub/Sub ไม่มี persistence — ถ้าไม่มี subscriber อยู่ตอน Publish ข้อความหาย; เหมาะกับ event แจ้งเตือน/audit ไม่ใช่คิวงาน critical |
| **Worker**      | ตรวจ PENDING หมดอายุแบบ polling (หรือตาม trigger) — ไม่ใช้ Redis Queue; ออกแบบให้เข้าใจ flow ปล่อยล็อกและส่ง event                     |
| **Scalability** | Backend รันหลายตัวได้เพราะ lock อยู่ที่ Redis; WebSocket ต้องใช้ sticky session หรือ adapter แชร์ room ระหว่าง instance                |
