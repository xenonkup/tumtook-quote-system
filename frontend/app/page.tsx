"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";

const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

// ตัวเลือกทั้งหมด — ต้องตรงกับ config ใน backend (pricing.go)
const PRODUCTS = [
  { id: "sticker", name: "สติกเกอร์ฉลาก" },
  { id: "box", name: "กล่องแพ็กเกจจิ้ง" },
  { id: "banner", name: "ป้าย/Banner Inkjet" },
];
const SIZES = [
  { id: "S", name: "S — เล็ก" },
  { id: "M", name: "M — กลาง" },
  { id: "L", name: "L — ใหญ่" },
];
const MATERIALS = [
  { id: "standard", name: "ธรรมดา" },
  { id: "waterproof", name: "กันน้ำ" },
  { id: "premium", name: "พรีเมียม" },
];
const OPTIONS = [
  { id: "lamination", name: "เคลือบเงา/ด้าน (+0.50 ฿/ชิ้น)" },
  { id: "diecut", name: "ไดคัทรูปทรง (+1.00 ฿/ชิ้น)" },
];

type QuoteResult = {
  id: number;
  productName: string;
  quantity: number;
  unitPrice: number;
  discountPercent: number;
  optionsCost: number;
  total: number;
};

const baht = (n: number) =>
  n.toLocaleString("th-TH", { style: "currency", currency: "THB" });

const fieldClass =
  "w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none transition-colors focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50";

export default function Home() {
  const [productId, setProductId] = useState("sticker");
  const [size, setSize] = useState("M");
  const [material, setMaterial] = useState("standard");
  const [quantity, setQuantity] = useState(100);
  const [options, setOptions] = useState<string[]>([]);
  const [customerName, setCustomerName] = useState("");
  const [customerPhone, setCustomerPhone] = useState("");

  const [result, setResult] = useState<QuoteResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const toggleOption = (id: string) =>
    setOptions((prev) =>
      prev.includes(id) ? prev.filter((o) => o !== id) : [...prev, id]
    );

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");
    setResult(null);
    try {
      const res = await fetch(`${API}/api/quote`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          productId,
          size,
          material,
          quantity: Number(quantity),
          options,
          customerName,
          customerPhone,
        }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.error ?? "เกิดข้อผิดพลาด");
        return;
      }
      setResult(data);
    } catch {
      setError("เชื่อมต่อ server ไม่ได้ — ตรวจว่า backend รันอยู่ที่ " + API);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="flex min-h-dvh items-center justify-center bg-muted/40 py-10 px-4">
      <div className="w-full max-w-4xl">
        {/* Header */}
        <header className="mb-8 text-center">
          <p className="text-sm font-medium text-muted-foreground">Tumtook · ทำถูก</p>
          <h1 className="mt-1 text-3xl font-semibold tracking-tight">
            คำนวณราคางานพิมพ์ออนไลน์
          </h1>
          <p className="mt-2 text-muted-foreground">
            กรอกสเปกงาน รับราคาประเมินทันที ไม่ต้องรอแอดมิน
          </p>
        </header>

        <div className="grid gap-6 md:grid-cols-5">
          {/* ===== Form ===== */}
          <form
            onSubmit={handleSubmit}
            className="rounded-xl border border-border bg-card p-6 shadow-sm md:col-span-3"
          >
            <div className="grid gap-5">
              <div className="grid gap-2">
                <label className="text-sm font-medium">ประเภทงาน</label>
                <select
                  className={fieldClass}
                  value={productId}
                  onChange={(e) => setProductId(e.target.value)}
                >
                  {PRODUCTS.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))}
                </select>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="grid gap-2">
                  <label className="text-sm font-medium">ขนาด</label>
                  <select
                    className={fieldClass}
                    value={size}
                    onChange={(e) => setSize(e.target.value)}
                  >
                    {SIZES.map((s) => (
                      <option key={s.id} value={s.id}>
                        {s.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="grid gap-2">
                  <label className="text-sm font-medium">วัสดุ</label>
                  <select
                    className={fieldClass}
                    value={material}
                    onChange={(e) => setMaterial(e.target.value)}
                  >
                    {MATERIALS.map((m) => (
                      <option key={m.id} value={m.id}>
                        {m.name}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="grid gap-2">
                <label className="text-sm font-medium">จำนวน (ชิ้น)</label>
                <input
                  type="number"
                  min={1}
                  className={fieldClass}
                  value={quantity}
                  onChange={(e) => setQuantity(Number(e.target.value))}
                />
                <p className="text-xs text-muted-foreground">
                  ยิ่งสั่งเยอะ ราคาต่อชิ้นยิ่งถูก (100+ ลด 8%, 500+ ลด 15%, 1000+ ลด 25%)
                </p>
              </div>

              <div className="grid gap-2">
                <label className="text-sm font-medium">ออปชันเสริม</label>
                <div className="grid gap-2">
                  {OPTIONS.map((o) => (
                    <label
                      key={o.id}
                      className="flex cursor-pointer items-center gap-2.5 rounded-lg border border-border px-3 py-2 text-sm hover:bg-muted/50"
                    >
                      <input
                        type="checkbox"
                        className="size-4 accent-primary"
                        checked={options.includes(o.id)}
                        onChange={() => toggleOption(o.id)}
                      />
                      {o.name}
                    </label>
                  ))}
                </div>
              </div>

              <hr className="border-border" />

              <div className="grid grid-cols-2 gap-4">
                <div className="grid gap-2">
                  <label className="text-sm font-medium">ชื่อผู้ติดต่อ</label>
                  <input
                    type="text"
                    placeholder="เช่น ร้านเบเกอรี่สุข"
                    className={fieldClass}
                    value={customerName}
                    onChange={(e) => setCustomerName(e.target.value)}
                  />
                </div>
                <div className="grid gap-2">
                  <label className="text-sm font-medium">เบอร์โทร</label>
                  <input
                    type="tel"
                    placeholder="08x-xxx-xxxx"
                    className={fieldClass}
                    value={customerPhone}
                    onChange={(e) => setCustomerPhone(e.target.value)}
                  />
                </div>
              </div>

              <Button type="submit" size="lg" disabled={loading} className="mt-1 w-full">
                {loading ? "กำลังคำนวณ…" : "คำนวณราคา"}
              </Button>

              {error && (
                <p className="rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
                  {error}
                </p>
              )}
            </div>
          </form>

          {/* ===== Result ===== */}
          <div className="md:col-span-2">
            <div className="sticky top-6 rounded-xl border border-border bg-card p-6 shadow-sm">
              <h2 className="text-sm font-medium text-muted-foreground">ราคาประเมิน</h2>

              {!result ? (
                <p className="mt-6 text-sm text-muted-foreground">
                  กรอกสเปกแล้วกด “คำนวณราคา” เพื่อดูราคาที่นี่
                </p>
              ) : (
                <div className="mt-3">
                  <p className="text-3xl font-semibold tracking-tight">
                    {baht(result.total)}
                  </p>
                  <p className="mt-1 text-sm text-muted-foreground">
                    {result.productName} · {result.quantity.toLocaleString()} ชิ้น
                  </p>

                  <dl className="mt-5 space-y-2 text-sm">
                    <div className="flex justify-between">
                      <dt className="text-muted-foreground">ราคา/ชิ้น</dt>
                      <dd>{baht(result.unitPrice)}</dd>
                    </div>
                    {result.discountPercent > 0 && (
                      <div className="flex justify-between text-emerald-600">
                        <dt>ส่วนลดจำนวน</dt>
                        <dd>−{result.discountPercent}%</dd>
                      </div>
                    )}
                    {result.optionsCost > 0 && (
                      <div className="flex justify-between">
                        <dt className="text-muted-foreground">ค่าออปชัน</dt>
                        <dd>{baht(result.optionsCost)}</dd>
                      </div>
                    )}
                    <div className="flex justify-between border-t border-border pt-2 font-medium">
                      <dt>รวมสุทธิ</dt>
                      <dd>{baht(result.total)}</dd>
                    </div>
                  </dl>

                  <p className="mt-5 rounded-lg bg-emerald-500/10 px-3 py-2 text-xs text-emerald-700">
                    ✓ บันทึกคำขอราคาแล้ว (เลขที่ #{result.id}) ทีมงานจะติดต่อกลับ
                  </p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </main>
  );
}
