package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"tumtook-quote/backend/internal/quote"
)

// Handler สร้าง mux + ลงทะเบียนทุก endpoint (ห่อด้วย CORS)
// main.go เรียกใช้ตัวนี้ตัวเดียว
func Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", withCORS(healthHandler))
	mux.HandleFunc("/api/quote", withCORS(quoteHandler))
	mux.HandleFunc("/api/quotes", withCORS(listQuotesHandler))
	return mux
}

// healthHandler ใช้เช็คว่า service ตื่นอยู่ (สำคัญกับ Render free tier ที่ sleep)
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// quoteHandler รับสเปกงาน -> คำนวณราคา -> บันทึก lead -> ตอบกลับ
func quoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "ใช้ POST เท่านั้น"})
		return
	}

	var req quote.QuoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "รูปแบบ JSON ไม่ถูกต้อง"})
		return
	}

	result, err := quote.Calculate(req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// บันทึก lead ลง DB (หัวใจของระบบ: ไม่พลาด lead)
	saved, err := quote.SaveQuote(r.Context(), req, result)
	if err != nil {
		log.Printf("บันทึก quote ไม่สำเร็จ: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "บันทึกข้อมูลไม่สำเร็จ"})
		return
	}

	writeJSON(w, http.StatusOK, saved)
}

// listQuotesHandler ดึง lead ล่าสุดสำหรับหน้าหลังบ้าน
func listQuotesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "ใช้ GET เท่านั้น"})
		return
	}

	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	quotes, err := quote.ListQuotes(r.Context(), limit)
	if err != nil {
		log.Printf("ดึง quotes ไม่สำเร็จ: %v", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "ดึงข้อมูลไม่สำเร็จ"})
		return
	}

	writeJSON(w, http.StatusOK, quotes)
}

// ===== helpers =====

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// withCORS เปิดให้ frontend (คนละ origin) เรียก API ได้
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}
