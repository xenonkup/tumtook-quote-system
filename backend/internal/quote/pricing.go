package quote

import (
	"fmt"
	"math"
)

// ===== Pricing config =====
// เก็บค่าราคาทั้งหมดไว้ที่เดียว ขั้นถัดไปจะย้ายไปเก็บใน Postgres
// เพื่อให้แก้ราคาได้โดยไม่ต้อง deploy ใหม่ (data-driven)

// Product คือประเภทสินค้า + ราคาฐานต่อชิ้น
type Product struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	BasePrice float64 `json:"basePrice"`
}

var products = map[string]Product{
	"sticker": {ID: "sticker", Name: "สติกเกอร์ฉลาก", BasePrice: 2.00},
	"box":     {ID: "box", Name: "กล่องแพ็กเกจจิ้ง", BasePrice: 4.20},
	"banner":  {ID: "banner", Name: "ป้าย/Banner Inkjet", BasePrice: 150.00},
}

var sizeFactors = map[string]float64{
	"S": 1.0,
	"M": 1.5,
	"L": 2.2,
}

var materialFactors = map[string]float64{
	"standard":   1.0, // ธรรมดา
	"waterproof": 1.3, // กันน้ำ
	"premium":    1.8, // พรีเมียม
}

// ค่าออปชันเสริม คิดเพิ่มต่อชิ้น
var optionCosts = map[string]float64{
	"lamination": 0.50, // เคลือบเงา/ด้าน
	"diecut":     1.00, // ไดคัทรูปทรง
}

const minimumOrder = 300.0 // ยอดสั่งขั้นต่ำ (บาท)

// ส่วนลดขั้นบันไดตามจำนวน (เศรษฐกิจขนาด)
func quantityDiscount(qty int) float64 {
	switch {
	case qty >= 1000:
		return 0.25
	case qty >= 500:
		return 0.15
	case qty >= 100:
		return 0.08
	default:
		return 0.0
	}
}

// ===== Request / Result =====

type QuoteRequest struct {
	ProductID string   `json:"productId"`
	Size      string   `json:"size"`
	Material  string   `json:"material"`
	Quantity  int      `json:"quantity"`
	Options   []string `json:"options"`

	// ข้อมูลลูกค้า (ใช้เก็บเป็น lead) — ไม่บังคับตอนคำนวณราคา
	CustomerName  string `json:"customerName"`
	CustomerPhone string `json:"customerPhone"`
}

type QuoteResult struct {
	ProductName     string  `json:"productName"`
	UnitPrice       float64 `json:"unitPrice"`
	DiscountPercent float64 `json:"discountPercent"`
	Subtotal        float64 `json:"subtotal"`
	OptionsCost     float64 `json:"optionsCost"`
	Total           float64 `json:"total"`
	MinimumApplied  bool    `json:"minimumApplied"`
	Currency        string  `json:"currency"`
}

// Calculate คือหัวใจของระบบ: รับสเปก -> คืนราคาประเมิน + รายละเอียด
func Calculate(req QuoteRequest) (QuoteResult, error) {
	p, ok := products[req.ProductID]
	if !ok {
		return QuoteResult{}, fmt.Errorf("ไม่รู้จักประเภทสินค้า: %q", req.ProductID)
	}
	sf, ok := sizeFactors[req.Size]
	if !ok {
		return QuoteResult{}, fmt.Errorf("ขนาดไม่ถูกต้อง: %q (รองรับ S, M, L)", req.Size)
	}
	mf, ok := materialFactors[req.Material]
	if !ok {
		return QuoteResult{}, fmt.Errorf("วัสดุไม่ถูกต้อง: %q", req.Material)
	}
	if req.Quantity < 1 {
		return QuoteResult{}, fmt.Errorf("จำนวนต้องมากกว่า 0")
	}

	unitPrice := p.BasePrice * sf * mf
	disc := quantityDiscount(req.Quantity)
	subtotal := unitPrice * float64(req.Quantity) * (1 - disc)

	var optionsCost float64
	for _, o := range req.Options {
		optionsCost += optionCosts[o] * float64(req.Quantity)
	}

	total := subtotal + optionsCost
	minApplied := false
	if total < minimumOrder {
		total = minimumOrder
		minApplied = true
	}

	return QuoteResult{
		ProductName:     p.Name,
		UnitPrice:       round2(unitPrice),
		DiscountPercent: disc * 100,
		Subtotal:        round2(subtotal),
		OptionsCost:     round2(optionsCost),
		Total:           round2(total),
		MinimumApplied:  minApplied,
		Currency:        "THB",
	}, nil
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
