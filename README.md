# SideLive

**Keep your live audience in sight.** SideLive adalah companion desktop lokal untuk membaca komentar TikTok LIVE dalam floating overlay, tanpa akun SideLive atau backend cloud.

![Status](https://img.shields.io/badge/status-MVP-8b5cf6) ![Go](https://img.shields.io/badge/Go-1.25-00ADD8) ![Wails](https://img.shields.io/badge/Wails-v3_beta-red) ![React](https://img.shields.io/badge/React-19-61DAFB)

## MVP yang tersedia

- TikTok LIVE provider melalui `PirateTok/live-go`, beserta demo provider untuk development tanpa live aktif.
- Lifecycle stream independen, status koneksi, pembatalan, bounded event fan-out, dan reconnect dengan exponential backoff + jitter.
- Normalisasi komentar menjadi domain event SideLive sebelum mencapai UI.
- Companion dashboard untuk menambah, menghubungkan, memutus, dan menghapus satu stream.
- Floating Wails window yang frameless, always-on-top, transparan pada Windows/macOS, dan click-through saat dikunci.
- Pengaturan opacity, font size, jumlah komentar, visibilitas, serta edit/lock mode.
- Konfigurasi JSON berversi dan atomik di direktori config user OS.
- Tidak ada akun, telemetry, database, atau server SideLive.

> TikTok adalah API tidak resmi dan dapat berubah. Demo provider tersedia dari pilihan platform untuk mengecek seluruh alur secara deterministik.

## Menjalankan development UI

Prasyarat: Go 1.25+, Node 20.20+, dan pnpm 10.

```bash
pnpm install
pnpm dev
```

Buka `http://localhost:5173`. Browser otomatis memakai bridge demo lokal. Pilih **Demo stream**, masukkan nama apa pun, lalu hubungkan untuk melihat komentar sintetis.

## Build desktop

Pastikan dependency native Wails untuk OS sudah terpasang. Kemudian:

```bash
pnpm build
rm -rf cmd/sidelive/frontend && mkdir -p cmd/sidelive/frontend
touch cmd/sidelive/frontend/.gitkeep
cp -r frontend/dist/. cmd/sidelive/frontend/
go build -o bin/sidelive ./cmd/sidelive
```

Jalankan `bin/sidelive`, pilih TikTok LIVE, masukkan username tanpa `@`, lalu klik **Hubungkan**. Buka tab Overlay untuk menampilkan dan mengunci overlay.

## Quality checks

```bash
go test ./...
go vet ./...
pnpm lint
pnpm typecheck
pnpm test
pnpm build
```

## Struktur

```text
cmd/sidelive       composition root + embedded production UI
internal/core      domain model netral-provider
internal/providers provider contract dan registry
internal/streams   lifecycle/reconnect per stream
internal/routing   bounded event fan-out
internal/config    local JSON persistence
internal/desktop   Wails overlay adapter
providers/tiktok  TikTok-native adapter + normalization
providers/demo    synthetic development provider
frontend           React companion + overlay presentation
```

Dokumen produk dan keputusan teknis lengkap tersedia di [`docs/PRD.md`](docs/PRD.md), [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md), dan [`docs/TECH_STACK.md`](docs/TECH_STACK.md).

## Batasan v0.1

MVP sengaja hanya mengekspos satu stream, satu overlay, dan event komentar. Transparansi overlay Linux bergantung compositor sehingga menggunakan solid fallback; click-through native Wails saat ini didukung Windows/macOS. Gift, multi-stream, history, filter, TTS, dan platform tambahan masuk roadmap setelah vertical slice ini.
