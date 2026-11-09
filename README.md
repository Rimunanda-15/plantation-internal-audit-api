# AACS Plantation Edition

**Audit & Assurance Control System — Backend API**

AACS Plantation Edition adalah sistem manajemen audit internal untuk perusahaan perkebunan. Backend dibangun menggunakan **Go**, **Fiber v3**, **MySQL**, dan pendekatan **Handler → Service → Repository**. Frontend direncanakan menggunakan **React + TypeScript**.

> Dokumen ini merupakan rekonstruksi spesifikasi dari pembahasan pengembangan sebelumnya. Daftar modul dan endpoint di bawah adalah cakupan rancangan; tidak semuanya berarti sudah diimplementasikan. Cocokkan dengan source code sebelum menandai fitur sebagai selesai.

## 1. Tujuan

- Mengelola SOP dan pengendalian dokumen.
- Mengidentifikasi, menilai, dan memantau risiko operasional.
- Menyusun rencana, program, dan pelaksanaan audit internal.
- Mendokumentasikan temuan, bukti audit, dan rekomendasi.
- Memantau corrective action serta tindak lanjut temuan.
- Menyediakan audit trail dan laporan untuk manajemen.
- Menggunakan data ERP sebagai sumber informasi audit tanpa mengubah transaksi ERP.

## 2. Teknologi

| Komponen | Teknologi |
| --- | --- |
| Bahasa | Go 1.25 |
| HTTP framework | Fiber v3 |
| Database | MySQL |
| Akses database | `sqlx` |
| Arsitektur | Handler / Controller → Service → Repository |
| Frontend (terpisah) | React + TypeScript |
| Development | Windows / PowerShell |
| Server target | Ubuntu 26.04.1 LTS |

## 3. Arsitektur dan integrasi ERP

Dua schema MySQL digunakan pada satu instance:

| Schema | Fungsi | Akses AACS |
| --- | --- | --- |
| `aacs_plantation` | Database utama AACS: SOP, risiko, audit, temuan, bukti, log | Read/write sesuai otorisasi |
| `plantation_antech` | Database ERP operasional | **Read-only melalui view yang disetujui** |

```text
React + TypeScript
        |
        v
AACS REST API (/api/v1)
        |
        v
Handler / Controller
        |
        v
Service (validasi dan aturan bisnis)
        |
        v
Repository (sqlx)
        |
        +----> aacs_plantation (read/write)
        |
        +----> ERP views (read-only)
                    |
                    v
              plantation_antech
```

**Aturan integrasi:**

1. AACS tidak melakukan `INSERT`, `UPDATE`, atau `DELETE` ke tabel transaksi ERP.
2. Akses data ERP dibatasi melalui view dan user database dengan hak `SELECT` seperlunya.
3. Hasil audit, penilaian, komentar, dan tindak lanjut disimpan di `aacs_plantation`.
4. Referensi ERP pada bukti/temuan harus mempertahankan identitas sumber untuk penelusuran.
5. Validasi izin akses dan audit trail dilakukan pada backend, bukan hanya frontend.

## 4. Cakupan modul

### Master Data

- Perusahaan, unit usaha, estate, mill, dan struktur organisasi.
- Departemen, pengguna, peran, dan hak akses.
- Kategori, status, prioritas, serta referensi audit.

### SOP & Document Control

- Kategori SOP dan daftar dokumen.
- Nomor dokumen, versi/revisi, tanggal berlaku, status aktif.
- Persetujuan dan riwayat perubahan dokumen.
- Lampiran serta hubungan SOP dengan risiko dan prosedur audit.

### Risk Management

- Risk register dan kategori risiko.
- Penilaian kemungkinan (likelihood) dan dampak (impact).
- Inherent risk, kontrol yang tersedia, residual risk.
- Risk owner, rencana mitigasi, target, dan monitoring.

### Audit Planning

- Audit universe dan rencana audit periodik.
- Penugasan auditor, auditee, ruang lingkup, jadwal, dan program audit.
- Penentuan prioritas berbasis risiko.

### Audit Execution

- Kertas kerja audit, checklist, prosedur pengujian.
- Sampel transaksi, observasi, dan dokumentasi pengujian.
- Referensi data ERP yang menjadi objek pemeriksaan.

### Audit Findings

- Temuan, kondisi, kriteria, sebab, akibat, dan rekomendasi.
- Klasifikasi tingkat risiko/keparahan temuan.
- Tanggapan auditee dan proses review temuan.

### Corrective Action & Follow-up

- Rencana aksi, PIC, tenggat, status, dan bukti penyelesaian.
- Verifikasi auditor, tindak lanjut, dan histori perubahan.
- Pemantauan temuan yang melewati tenggat.

### Evidence & Audit Trail

- Metadata bukti, lampiran, referensi transaksi, dan jejak aktivitas.
- Pencatatan pelaku, waktu, jenis tindakan, dan objek perubahan.
- Pengendalian akses dan integritas file bukti.

### Reporting

- Ringkasan status audit, risiko, temuan, dan tindak lanjut.
- Ekspor laporan ke **Excel, PDF, DOCX, dan PPTX** sesuai kebutuhan modul.
- **HTML bukan format keluaran laporan** yang direncanakan.

## 5. Struktur project

Struktur konseptual berikut mengikuti pembagian package yang dibahas; sesuaikan nama file/folder dengan implementasi repository aktual.

```text
aacs-api/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── config/
│   ├── database/
│   ├── middleware/
│   ├── models/
│   ├── controllers/
│   ├── services/
│   ├── repositories/
│   ├── routes/
│   └── modules/
├── .env
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

Tanggung jawab lapisan:

- **Routes:** registrasi endpoint dan middleware.
- **Controller/Handler:** membaca request, validasi bentuk input, mengembalikan respons HTTP.
- **Service:** aturan bisnis, workflow, dan transaksi aplikasi.
- **Repository:** query SQL serta pemetaan data.
- **Models:** struktur domain, request/response, dan metadata data.
- **Middleware:** autentikasi, otorisasi, logging, dan proteksi request.

## 6. Persiapan development

Prasyarat:

- Go 1.25 atau versi yang kompatibel dengan `go.mod`.
- MySQL yang dapat diakses dari komputer development.
- Git.
- Database `aacs_plantation` dan akses read-only ke view ERP yang dibutuhkan.

### Clone dan install dependency

```powershell
git clone <URL_REPOSITORY_AACS>
cd aacs-api
go mod download
```

Jika repository sudah tersedia:

```powershell
cd D:\Antech\AACS\aacs-api
go mod tidy
```

### Konfigurasi environment

Buat file `.env` pada root project. Contoh untuk development lokal:

```dotenv
APP_NAME=AACS Plantation API
APP_PORT=8081
APP_ENV=development

DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=
DB_NAME=aacs_plantation
```

Nilai di atas hanya contoh lokal, **bukan konfigurasi production**. Gunakan user database khusus dengan hak akses minimum. Jangan commit `.env` atau credential ke Git.

Jika koneksi ke schema ERP menggunakan kredensial berbeda, definisikan variabel konfigurasi dan pool koneksi terpisah pada implementasi; jangan mengasumsikan bahwa konfigurasi ERP tersebut sudah tersedia.

### Menjalankan aplikasi

```powershell
go run ./cmd/api
```

Periksa health endpoint yang sudah dibahas pada pengembangan sebelumnya:

```powershell
Invoke-RestMethod http://localhost:8081/api/v1/health
```

Health endpoint: `GET /api/v1/health`.

### Build

```powershell
go build -o aacs-api.exe ./cmd/api
.\aacs-api.exe
```

Untuk build Linux dari PowerShell:

```powershell
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -o aacs-api ./cmd/api
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
```

## 7. Konvensi API

Base URL development:

```text
http://localhost:8081/api/v1
```

Endpoint yang telah disebut dalam pembahasan:

| Area | Path | Keterangan |
| --- | --- | --- |
| Health | `GET /api/v1/health` | Pemeriksaan layanan |
| SOP | `/api/v1/sop/*` | Kelompok endpoint SOP |
| Reports | `/api/v1/reports/*` | Kelompok endpoint laporan |

> Route SOP dan laporan di atas merupakan pola/kelompok route yang dibahas, **bukan daftar endpoint final**. Metode HTTP, nama path detail, payload, dan status implementasi harus diverifikasi di package `routes` pada source code.

### Prinsip respons dan validasi

- Gunakan HTTP status code yang sesuai.
- Validasi input di boundary request dan aturan bisnis di service.
- Hindari mengirim detail SQL, stack trace, atau credential ke client.
- Gunakan pagination dan filter untuk daftar dengan volume data besar.
- Terapkan pemeriksaan hak akses pada setiap operasi yang sensitif.

## 8. Keamanan

- Autentikasi dan otorisasi berbasis role/permission.
- Principle of least privilege untuk akun database.
- Akun ERP hanya diberi `SELECT` pada view yang disetujui.
- Password/secret tidak ditulis langsung di source code.
- Gunakan HTTPS di production melalui reverse proxy.
- Validasi file upload, ukuran, ekstensi, dan akses bukti audit.
- Simpan audit trail untuk perubahan data penting.
- Terapkan parameterized queries dan transaksi database bila diperlukan.

## 9. Testing dan kualitas kode

```powershell
go fmt ./...
go vet ./...
go test ./...
```

Untuk memeriksa dependency:

```powershell
go list -m all
```

Sebelum merge, pastikan perubahan tidak merusak kontrak API, aturan otorisasi, integrasi read-only ERP, dan migrasi database.

## 10. Deployment Ubuntu

Target server yang dibahas: **Ubuntu 26.04.1 LTS**.

Alur deployment yang disarankan:

1. Siapkan binary Linux sesuai arsitektur server.
2. Siapkan file environment production di luar Git.
3. Gunakan user sistem non-root untuk menjalankan API.
4. Jalankan service dengan `systemd` atau orkestrasi deployment yang dipilih.
5. Pasang reverse proxy dan HTTPS.
6. Batasi akses port API serta database.
7. Konfigurasikan log, backup schema AACS, dan monitoring.
8. Uji health endpoint dan izin akses ERP read-only.

Contoh binary build di server:

```bash
go build -o aacs-api ./cmd/api
./aacs-api
```

## 11. Status pengembangan

Informasi yang telah dikonfirmasi dalam pembahasan:

- [x] Project backend Go berjalan dan terhubung ke MySQL.
- [x] Endpoint health `GET /api/v1/health` telah dibahas sebagai endpoint yang bekerja.
- [x] Pemisahan schema AACS dan ERP ditetapkan sebagai prinsip arsitektur.
- [ ] Verifikasi implementasi aktual seluruh endpoint SOP.
- [ ] Verifikasi implementasi aktual risk register dan workflow audit.
- [ ] Verifikasi endpoint findings, corrective actions, evidence, dan reports.
- [ ] Lengkapi dokumentasi payload, migrasi database, test, dan deployment production.

Checklist yang belum tercentang berarti **belum diverifikasi dari repository**, bukan pernyataan bahwa fitur tersebut pasti belum dibuat.

## 12. Prinsip pengembangan

1. **ERP tetap menjadi sumber data operasional**, bukan tempat menyimpan keputusan audit AACS.
2. **AACS menjadi sumber data audit**, termasuk SOP, risiko, temuan, rekomendasi, dan tindak lanjut.
3. **Setiap temuan dapat ditelusuri** ke kriteria, bukti, auditor, auditee, dan sumber transaksi bila relevan.
4. **Workflow audit harus dapat diaudit**, termasuk perubahan status dan persetujuan.
5. **Backend menjadi penegak aturan bisnis dan otorisasi**.

---

**AACS Plantation Edition — Audit & Assurance Control System**
