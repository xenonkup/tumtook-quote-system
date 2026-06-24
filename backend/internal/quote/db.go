package quote

import (
	"context"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// dbpool คือ connection pool ที่หมุนเวียนใช้ทุก request (best practice สำหรับ web API)
var dbpool *pgxpool.Pool

// connString อ่านจาก env DATABASE_URL ถ้าไม่มีใช้ค่า dev (Docker local) เป็น default
func connString() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://postgres:devpass@localhost:5432/tumtook"
}

// InitDB เปิด pool + สร้างตารางถ้ายังไม่มี (auto-migrate ง่ายๆ)
func InitDB(ctx context.Context) error {
	pool, err := pgxpool.New(ctx, connString())
	if err != nil {
		return err
	}

	// เช็คว่าต่อ DB ได้จริงก่อนไปต่อ
	if err := pool.Ping(ctx); err != nil {
		return err
	}

	dbpool = pool
	return createSchema(ctx)
}

// CloseDB ปิด connection pool (เรียกตอน shutdown)
func CloseDB() {
	if dbpool != nil {
		dbpool.Close()
	}
}

func createSchema(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS quotes (
    id             BIGSERIAL PRIMARY KEY,
    product_id     TEXT          NOT NULL,
    product_name   TEXT          NOT NULL,
    size           TEXT          NOT NULL,
    material       TEXT          NOT NULL,
    quantity       INTEGER       NOT NULL,
    options        TEXT[]        NOT NULL DEFAULT '{}',
    unit_price     NUMERIC(10,2) NOT NULL,
    discount_pct   NUMERIC(5,2)  NOT NULL,
    options_cost   NUMERIC(10,2) NOT NULL,
    total          NUMERIC(10,2) NOT NULL,
    customer_name  TEXT,
    customer_phone TEXT,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now()
);`
	_, err := dbpool.Exec(ctx, schema)
	return err
}

// Quote คือ lead ที่บันทึกแล้ว (มี id + created_at เพิ่มจากผลคำนวณ)
type Quote struct {
	ID            int64     `json:"id"`
	ProductID     string    `json:"productId"`
	ProductName   string    `json:"productName"`
	Size          string    `json:"size"`
	Material      string    `json:"material"`
	Quantity      int       `json:"quantity"`
	Options       []string  `json:"options"`
	UnitPrice     float64   `json:"unitPrice"`
	DiscountPct   float64   `json:"discountPercent"`
	OptionsCost   float64   `json:"optionsCost"`
	Total         float64   `json:"total"`
	CustomerName  string    `json:"customerName"`
	CustomerPhone string    `json:"customerPhone"`
	CreatedAt     time.Time `json:"createdAt"`
}

// SaveQuote บันทึก lead ลง DB แล้วคืน id + created_at
func SaveQuote(ctx context.Context, req QuoteRequest, res QuoteResult) (Quote, error) {
	q := Quote{
		ProductID:     req.ProductID,
		ProductName:   res.ProductName,
		Size:          req.Size,
		Material:      req.Material,
		Quantity:      req.Quantity,
		Options:       req.Options,
		UnitPrice:     res.UnitPrice,
		DiscountPct:   res.DiscountPercent,
		OptionsCost:   res.OptionsCost,
		Total:         res.Total,
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
	}
	if q.Options == nil {
		q.Options = []string{}
	}

	const insert = `
INSERT INTO quotes
    (product_id, product_name, size, material, quantity, options,
     unit_price, discount_pct, options_cost, total, customer_name, customer_phone)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
RETURNING id, created_at;`

	err := dbpool.QueryRow(ctx, insert,
		q.ProductID, q.ProductName, q.Size, q.Material, q.Quantity, q.Options,
		q.UnitPrice, q.DiscountPct, q.OptionsCost, q.Total, q.CustomerName, q.CustomerPhone,
	).Scan(&q.ID, &q.CreatedAt)

	return q, err
}

// ListQuotes ดึง lead ล่าสุด (สำหรับหน้าหลังบ้าน) จำกัด limit แถว
func ListQuotes(ctx context.Context, limit int) ([]Quote, error) {
	const query = `
SELECT id, product_id, product_name, size, material, quantity, options,
       unit_price, discount_pct, options_cost, total, customer_name, customer_phone, created_at
FROM quotes
ORDER BY created_at DESC
LIMIT $1;`

	rows, err := dbpool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	quotes := []Quote{}
	for rows.Next() {
		var q Quote
		if err := rows.Scan(
			&q.ID, &q.ProductID, &q.ProductName, &q.Size, &q.Material, &q.Quantity, &q.Options,
			&q.UnitPrice, &q.DiscountPct, &q.OptionsCost, &q.Total, &q.CustomerName, &q.CustomerPhone, &q.CreatedAt,
		); err != nil {
			return nil, err
		}
		quotes = append(quotes, q)
	}
	return quotes, rows.Err()
}
