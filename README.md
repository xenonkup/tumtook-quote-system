# Tumtook Instant Quote — เครื่องคำนวณราคางานพิมพ์ออนไลน์

ระบบประเมินราคางานพิมพ์แบบเรียลไทม์สำหรับ **Tumtook (ทำถูก)** — ลูกค้ากรอกสเปกงาน → ได้ราคาทันที → ระบบบันทึกเป็น *lead* เข้าหลังบ้านอัตโนมัติ

> **โจทย์ที่แก้:** Tumtook เป็นโรงพิมพ์/แพ็กเกจจิ้งที่รับงาน custom จำนวนมาก ลูกค้าต้องทักแชทถามราคาแล้วรอแอดมินตอบ ทำให้เสีย lead นอกเวลาทำการ — ระบบนี้ตอบราคาได้ **24 ชม. และไม่พลาด lead**

---

## 📑 สารบัญ
- [ฟีเจอร์](#-ฟีเจอร์)
- [Tech Stack](#-tech-stack)
- [โครงสร้างโปรเจค](#-โครงสร้างโปรเจค)
- [วิธีคำนวณราคา](#-วิธีคำนวณราคา)
- [API Endpoints](#-api-endpoints)
- [Database](#-database)
- [การติดตั้งและรัน](#-การติดตั้งและรัน)
- [Environment Variables](#-environment-variables)
- [การ Deploy](#-การ-deploy)
- [การตัดสินใจทางเทคนิค & การใช้ AI](#-การตัดสินใจทางเทคนิค--การใช้-ai)

---

## ✨ ฟีเจอร์
- กรอกสเปกงาน (ประเภท → ขนาด → วัสดุ → จำนวน → ออปชัน) แล้วได้ราคาประเมินทันที พร้อม breakdown
- ส่วนลดตามจำนวน (เศรษฐกิจขนาด) + ยอดสั่งขั้นต่ำ — เหมือนการคิดราคาจริงของโรงพิมพ์
- บันทึกทุกคำขอราคาเป็น lead (ชื่อ/เบอร์ลูกค้า) ลงฐานข้อมูล
- endpoint สำหรับดึง lead ไปแสดงหน้าหลังบ้าน

---

## 🛠 Tech Stack

| Layer | เทคโนโลยี | Deploy |
|---|---|---|
| Frontend | Next.js 16 · TypeScript · Tailwind CSS v4 · shadcn/ui | Vercel |
| Backend | Go (stdlib `net/http`) | Render (Web Service) |
| Database | PostgreSQL · pgx / pgxpool | Render (Managed Postgres) |
| Dev DB | PostgreSQL บน Docker | local |

**สถาปัตยกรรม:**
```
[ผู้ใช้] → Vercel (Next.js) → HTTP → Render (Go API) → Render (PostgreSQL)
```

---

## 📁 โครงสร้างโปรเจค

```
tumtook-quote/
├── backend/
│   ├── go.mod / go.sum
│   ├── .env.example
│   ├── server/
│   │   └── main.go            # bootstrap: เปิด DB + start HTTP server
│   ├── routes/
│   │   └── routes.go          # HTTP layer: routes + handlers + CORS
│   └── internal/quote/
│       ├── pricing.go         # business logic: สูตรคิดราคา + config
│       └── db.go              # data layer: Postgres pool + บันทึก/ดึง lead
└── frontend/
    ├── .env.local             # ค่าจริง (gitignore)
    ├── .env.example
    ├── app/
    │   ├── layout.tsx
    │   └── page.tsx           # หน้าฟอร์มขอราคา
    └── lib/
        └── fetchdata.ts       # API client (createQuote / listQuotes)
```

แต่ละ layer แยกหน้าที่ชัดเจน: `server` (bootstrap) → `routes` (HTTP) → `quote` (logic + DB)

---

## 🧮 วิธีคำนวณราคา

หัวใจของระบบอยู่ที่ [`backend/internal/quote/pricing.go`](backend/internal/quote/pricing.go)

### สูตร
```
unitPrice = basePrice × sizeFactor × materialFactor
subtotal  = unitPrice × quantity × (1 − quantityDiscount)
total     = max(subtotal + optionsCost, minimumOrder)
```
โดย `optionsCost = Σ (ค่าออปชันแต่ละตัว × quantity)`

### ตารางค่าที่ใช้ (เก็บเป็น config — ปรับราคาได้โดยไม่ต้องแก้สูตร)

**ราคาฐานต่อชิ้น (basePrice)**
| ประเภทงาน | รหัส | ราคา |
|---|---|---|
| สติกเกอร์ฉลาก | `sticker` | 2.00 ฿ |
| กล่องแพ็กเกจจิ้ง | `box` | 4.20 ฿ |
| ป้าย/Banner Inkjet | `banner` | 150.00 ฿ |

**ตัวคูณขนาด (sizeFactor):** `S` = 1.0 · `M` = 1.5 · `L` = 2.2
**ตัวคูณวัสดุ (materialFactor):** `standard` (ธรรมดา) = 1.0 · `waterproof` (กันน้ำ) = 1.3 · `premium` (พรีเมียม) = 1.8

**ส่วนลดตามจำนวน (quantityDiscount)**
| จำนวน (ชิ้น) | ส่วนลด |
|---|---|
| 1 – 99 | 0% |
| 100 – 499 | 8% |
| 500 – 999 | 15% |
| 1,000 ขึ้นไป | 25% |

**ค่าออปชันเสริม (ต่อชิ้น):** `lamination` (เคลือบเงา/ด้าน) = +0.50 ฿ · `diecut` (ไดคัทรูปทรง) = +1.00 ฿
**ยอดสั่งขั้นต่ำ (minimumOrder):** 300 ฿

### ตัวอย่าง
> สติกเกอร์ฉลาก · ขนาด M · กันน้ำ · 500 ชิ้น · เคลือบเงา
```
unitPrice = 2.00 × 1.5 × 1.3        = 3.90 ฿/ชิ้น
subtotal  = 3.90 × 500 × (1 − 0.15) = 1,657.50 ฿
total     = 1,657.50 + (0.50 × 500) = 1,907.50 ฿   (เกินขั้นต่ำ 300)
```

---

## 🔌 API Endpoints

Base URL (dev): `http://localhost:8080`

### `GET /health`
เช็คว่า service ตื่นอยู่ (ใช้กับ Render free tier ที่ sleep)
```json
{ "status": "ok" }
```

### `POST /api/quote`
คำนวณราคา + บันทึก lead → คืนข้อมูลที่บันทึกพร้อม `id`

**Request body**
```json
{
  "productId": "sticker",
  "size": "M",
  "material": "waterproof",
  "quantity": 500,
  "options": ["lamination"],
  "customerName": "ร้านเบเกอรี่สุข",
  "customerPhone": "0812345678"
}
```

**Response `200`**
```json
{
  "id": 1,
  "productId": "sticker",
  "productName": "สติกเกอร์ฉลาก",
  "size": "M",
  "material": "waterproof",
  "quantity": 500,
  "options": ["lamination"],
  "unitPrice": 3.9,
  "discountPercent": 15,
  "optionsCost": 250,
  "total": 1907.5,
  "customerName": "ร้านเบเกอรี่สุข",
  "customerPhone": "0812345678",
  "createdAt": "2026-06-24T17:15:05+07:00"
}
```
**Error `400`** เมื่อสเปกไม่ถูกต้อง: `{ "error": "ขนาดไม่ถูกต้อง: \"XL\" (รองรับ S, M, L)" }`

### `GET /api/quotes?limit=N`
ดึง lead ล่าสุด (เรียงใหม่→เก่า, `limit` 1–200 ค่า default 50) — คืน array ของ object แบบเดียวกับด้านบน

---

## 🗄 Database

ตาราง `quotes` ถูกสร้างอัตโนมัติตอน start (`CREATE TABLE IF NOT EXISTS`)

| คอลัมน์ | ชนิด | หมายเหตุ |
|---|---|---|
| `id` | BIGSERIAL | PK |
| `product_id` / `product_name` | TEXT | |
| `size` / `material` | TEXT | |
| `quantity` | INTEGER | |
| `options` | TEXT[] | array ของออปชัน |
| `unit_price` / `options_cost` / `total` | NUMERIC(10,2) | |
| `discount_pct` | NUMERIC(5,2) | |
| `customer_name` / `customer_phone` | TEXT | ข้อมูล lead |
| `created_at` | TIMESTAMPTZ | default `now()` |

---

## 🚀 การติดตั้งและรัน

### สิ่งที่ต้องมี
- Go 1.21+ · Node.js 18+ · Docker (สำหรับ Postgres ตอน dev)

### 1) Database (Postgres ผ่าน Docker)
```bash
docker run --name tumtook-pg -e POSTGRES_PASSWORD=devpass -e POSTGRES_DB=tumtook -p 5432:5432 -d postgres:16
# เปิดใหม่ภายหลัง: docker start tumtook-pg
```

### 2) Backend (Go)
```bash
cd backend
go run ./server          # ⚠️ ต้องเป็น ./server (ไม่ใช่ main.go)
# → Tumtook Quote API running on :8080
```

### 3) Frontend (Next.js)
```bash
cd frontend
cp .env.example .env.local   # ตั้ง NEXT_PUBLIC_API_URL (default = http://localhost:8080)
npm install
npm run dev                  # → http://localhost:3000
```

### ทดสอบ API เร็วๆ
```bash
curl http://localhost:8080/health
curl -X POST http://localhost:8080/api/quote \
  -H "Content-Type: application/json" \
  -d '{"productId":"sticker","size":"M","material":"waterproof","quantity":500,"options":["lamination"]}'
```

---

## 🔑 Environment Variables

| ตัวแปร | ที่ใช้ | dev | production |
|---|---|---|---|
| `NEXT_PUBLIC_API_URL` | frontend | `http://localhost:8080` | URL ของ backend บน Render |
| `DATABASE_URL` | backend | มี default (Docker local) ในโค้ด | Postgres URL ของ Render |
| `PORT` | backend | `8080` | Render กำหนดให้เอง |

> ⚠️ Next.js อ่าน env **ตอน start เท่านั้น** — แก้ `.env.local` ต้อง restart dev server | `NEXT_PUBLIC_*` ถูก inline ตอน build ดังนั้นบน Vercel ต้องตั้งค่า **ก่อน build**
> ไฟล์ที่มีค่าจริง (`.env.local`) ถูก gitignore — ใช้ `.env.example` เป็น template

---

## ☁️ การ Deploy

1. **Render → PostgreSQL** — สร้าง Managed Postgres → copy *Internal Database URL*
2. **Render → Web Service (backend)** — Root `backend/` · Build `go build -o app ./server` · Start `./app` · ตั้ง env `DATABASE_URL`
3. **Vercel (frontend)** — Root `frontend/` · ตั้ง env `NEXT_PUBLIC_API_URL` = URL ของ backend บน Render

> หมายเหตุ Render free tier: service จะ *sleep* หลังไม่มี request ~15 นาที (ตื่นเองเมื่อมีคนเข้า ~30 วิ) · Postgres ฟรีมีอายุ 30 วัน

---

## 🧠 การตัดสินใจทางเทคนิค & การใช้ AI

### เหตุผลการเลือกเทคโนโลยี
| เลือก | แทน | เหตุผล |
|---|---|---|
| Go stdlib `net/http` | Gin / Fiber | scope เล็ก ไม่ต้องพึ่ง framework ลด dependency, deploy ง่าย |
| pgx + pgxpool | lib/pq | performance ดีกว่า + connection pool เป็น best practice สำหรับ web API |
| PostgreSQL | MySQL | Render มี managed Postgres ฟรี อยู่ data center เดียวกับ backend → latency ต่ำ |
| Monorepo | 2 repos แยก | ส่ง GitHub ลิงก์เดียว ดูง่าย |
| Config ราคาเป็น data | hardcode ในสูตร | แก้ราคาได้โดยไม่ต้องแก้ logic (เตรียมย้ายไปเก็บใน DB ต่อได้) |
| แยก package `routes` / `quote` | รวมใน main | แยก HTTP layer ออกจาก business logic → test แยกได้, โตต่อง่าย |

### การใช้ AI ช่วยพัฒนา
ใช้ **Claude (AI assistant)** ช่วยในส่วน:
- วิเคราะห์ธุรกิจ Tumtook จากข้อมูลสาธารณะ และเลือกฟีเจอร์ที่ตรง pain point
- ออกแบบสูตรคิดราคา + schema ฐานข้อมูล
- เขียนโค้ด Go (API + pricing + DB) และ Next.js (ฟอร์ม + API client) พร้อมตรวจสอบการทำงานจริง (verify ครบ flow)
- ตั้งค่าสภาพแวดล้อม (Docker, WSL2, Git) และ troubleshoot ปัญหา

ทุกการตัดสินใจทางสถาปัตยกรรมและ trade-off ผ่านการพิจารณาและยืนยันร่วมกัน ไม่ใช่รับโค้ดมาทั้งก้อนโดยไม่เข้าใจ
