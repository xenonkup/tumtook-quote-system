// API client — รวม logic เรียก Go backend ไว้ที่เดียว
// component อื่นแค่ import ฟังก์ชันจากไฟล์นี้ ไม่ต้องเขียน fetch เอง

export const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// ===== Types (ตรงกับ backend) =====

export type QuoteRequest = {
  productId: string;
  size: string;
  material: string;
  quantity: number;
  options: string[];
  customerName?: string;
  customerPhone?: string;
};

// Quote = lead ที่บันทึกแล้ว (ค่าที่ POST /api/quote คืนกลับมา)
export type Quote = {
  id: number;
  productId: string;
  productName: string;
  size: string;
  material: string;
  quantity: number;
  options: string[];
  unitPrice: number;
  discountPercent: number;
  optionsCost: number;
  total: number;
  customerName: string;
  customerPhone: string;
  createdAt: string;
};

// request helper — จัดการ network error + error จาก API ที่เดียว
async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${API_BASE}${path}`, init);
  } catch {
    throw new Error(`เชื่อมต่อ server ไม่ได้ — ตรวจว่า backend รันอยู่ที่ ${API_BASE}`);
  }

  const data = await res.json().catch(() => null);
  if (!res.ok) {
    throw new Error(data?.error ?? "เกิดข้อผิดพลาด");
  }
  return data as T;
}

// ===== API calls =====

// createQuote: คำนวณราคา + บันทึก lead
export function createQuote(req: QuoteRequest): Promise<Quote> {
  return request<Quote>("/api/quote", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(req),
  });
}

// listQuotes: ดึง lead ล่าสุด (สำหรับหน้าหลังบ้าน)
export function listQuotes(limit = 50): Promise<Quote[]> {
  return request<Quote[]>(`/api/quotes?limit=${limit}`);
}
