# P0 單位圖 — 尺寸閘（128×128）

- **畫布**：輸出檔 MUST 為 exactly **128×128** pixels（生成後以 `identify` 或 PIL 驗證；不符則 Nearest-neighbor 裁切／縮放至 128×128，仍不可用則重製）。
- **剪影高度**：主體 silhouette 高度 **96–128** px（留足上下邊距，一圖一主體）。
- **背景／風格**：與角色卡同一 cartoon-pixel 規範時，背景 `#1A1410`、硬 1px 描邊、左上光源；詳見 `p0_units_v02.md` 路徑與調色。
