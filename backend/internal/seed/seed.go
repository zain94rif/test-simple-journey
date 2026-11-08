package seed

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"helpdesk/backend/internal/auth"
	"helpdesk/backend/internal/config"
	"helpdesk/backend/internal/domain"
)

type account struct {
	config config.SeedAccount
	role   domain.Role
}

type demoComment struct {
	authorEmail string
	body        string
	isInternal  bool
	// minutesAfterCreation keeps comment timestamps ordered and realistic.
	minutesAfterCreation int
}

type demoTicket struct {
	title          string
	description    string
	category       domain.Category
	priority       domain.Priority
	status         domain.Status
	requesterEmail string
	assigneeEmail  string
	// createdHoursAgo positions the ticket in the list so sorting and
	// pagination look meaningful during manual testing.
	createdHoursAgo int
	comments        []demoComment
}

// Run inserts the demo accounts (and, when enabled, demo tickets/comments) if
// they do not exist. Existing rows are never modified, so repeated runs are
// safe and fast.
func Run(ctx context.Context, pool *pgxpool.Pool, cfg config.Config) error {
	accounts := []account{
		{config: cfg.SeedAdmin, role: domain.RoleAdmin},
		{config: cfg.SeedAgent, role: domain.RoleAgent},
		{config: cfg.SeedAgent2, role: domain.RoleAgent},
		{config: cfg.SeedRequester, role: domain.RoleRequester},
		{config: cfg.SeedRequester2, role: domain.RoleRequester},
		{config: cfg.SeedRequester3, role: domain.RoleRequester},
	}

	userIDs := make(map[string]int64, len(accounts))
	for _, acc := range accounts {
		id, err := upsertAccount(ctx, pool, acc, cfg.BcryptCost)
		if err != nil {
			return err
		}
		userIDs[acc.config.Email] = id
	}

	if !cfg.SeedDemoData {
		return nil
	}
	return seedDemoTickets(ctx, pool, userIDs)
}

// upsertAccount returns the user ID, hashing the password only when the
// account has to be created.
func upsertAccount(ctx context.Context, pool *pgxpool.Pool, acc account, bcryptCost int) (int64, error) {
	var id int64
	err := pool.QueryRow(ctx,
		`SELECT id FROM users WHERE email = lower($1)`, acc.config.Email,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("cek seed %s: %w", acc.role, err)
	}

	hash, err := auth.HashPassword(acc.config.Password, bcryptCost)
	if err != nil {
		return 0, fmt.Errorf("hash password seed %s: %w", acc.role, err)
	}
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (name, email, password_hash, role)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
		RETURNING id`,
		acc.config.Name, acc.config.Email, hash, string(acc.role),
	).Scan(&id); err != nil {
		// A concurrent seeder may have inserted the row after our SELECT.
		if err := pool.QueryRow(ctx,
			`SELECT id FROM users WHERE email = lower($1)`, acc.config.Email,
		).Scan(&id); err != nil {
			return 0, fmt.Errorf("seed %s: %w", acc.role, err)
		}
	}
	return id, nil
}

func seedDemoTickets(ctx context.Context, pool *pgxpool.Pool, userIDs map[string]int64) error {
	now := time.Now().UTC()
	for _, demo := range demoTickets() {
		requesterID, ok := userIDs[demo.requesterEmail]
		if !ok {
			return fmt.Errorf("seed tiket %q: requester %s tidak dikenal", demo.title, demo.requesterEmail)
		}
		var assigneeID *int64
		if demo.assigneeEmail != "" {
			id, ok := userIDs[demo.assigneeEmail]
			if !ok {
				return fmt.Errorf("seed tiket %q: assignee %s tidak dikenal", demo.title, demo.assigneeEmail)
			}
			assigneeID = &id
		}

		createdAt := now.Add(-time.Duration(demo.createdHoursAgo) * time.Hour)
		updatedAt := createdAt
		if len(demo.comments) > 0 {
			updatedAt = createdAt.Add(time.Duration(demo.comments[len(demo.comments)-1].minutesAfterCreation) * time.Minute)
		}

		ticketID, err := ensureTicket(ctx, pool, demo, requesterID, assigneeID, createdAt, updatedAt)
		if err != nil {
			return err
		}

		for _, comment := range demo.comments {
			authorID, ok := userIDs[comment.authorEmail]
			if !ok {
				return fmt.Errorf("seed komentar tiket %q: author %s tidak dikenal", demo.title, comment.authorEmail)
			}
			if err := ensureComment(ctx, pool, ticketID, authorID, comment, createdAt); err != nil {
				return err
			}
		}
	}
	return nil
}

func ensureTicket(
	ctx context.Context,
	pool *pgxpool.Pool,
	demo demoTicket,
	requesterID int64,
	assigneeID *int64,
	createdAt, updatedAt time.Time,
) (int64, error) {
	var id int64
	err := pool.QueryRow(ctx, `
		SELECT id FROM tickets
		WHERE requester_id = $1 AND title = $2`,
		requesterID, demo.title,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("cek tiket %q: %w", demo.title, err)
	}

	if err := pool.QueryRow(ctx, `
		INSERT INTO tickets
			(title, description, category, priority, status, requester_id, assignee_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		demo.title, demo.description, string(demo.category), string(demo.priority),
		string(demo.status), requesterID, assigneeID, createdAt, updatedAt,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("seed tiket %q: %w", demo.title, err)
	}
	return id, nil
}

func ensureComment(
	ctx context.Context,
	pool *pgxpool.Pool,
	ticketID, authorID int64,
	comment demoComment,
	ticketCreatedAt time.Time,
) error {
	body := strings.TrimSpace(comment.body)
	var exists bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM comments WHERE ticket_id = $1 AND author_id = $2 AND body = $3
		)`, ticketID, authorID, body,
	).Scan(&exists); err != nil {
		return fmt.Errorf("cek komentar tiket %d: %w", ticketID, err)
	}
	if exists {
		return nil
	}

	createdAt := ticketCreatedAt.Add(time.Duration(comment.minutesAfterCreation) * time.Minute)
	if _, err := pool.Exec(ctx, `
		INSERT INTO comments (ticket_id, author_id, body, is_internal, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		ticketID, authorID, body, comment.isInternal, createdAt,
	); err != nil {
		return fmt.Errorf("seed komentar tiket %d: %w", ticketID, err)
	}
	return nil
}

// demoTickets covers every status, priority, and category so filters, sorting,
// role scoping, and the internal-note rules can be exercised manually.
func demoTickets() []demoTicket {
	return []demoTicket{
		{
			title:           "Laptop tidak bisa booting setelah update Windows",
			description:     "Setelah update Windows tadi malam, laptop berhenti di layar logo dan tidak mau masuk ke desktop.",
			category:        domain.CategoryHardware,
			priority:        domain.PriorityHigh,
			status:          domain.StatusInProgress,
			requesterEmail:  "requester@helpdesk.local",
			assigneeEmail:   "agent@helpdesk.local",
			createdHoursAgo: 26,
			comments: []demoComment{
				{authorEmail: "requester@helpdesk.local", body: "Sudah saya coba restart tiga kali, tetap berhenti di logo Windows.", minutesAfterCreation: 10},
				{authorEmail: "agent@helpdesk.local", body: "Sedang saya siapkan media recovery. Jangan lupa backup data penting dulu.", isInternal: false, minutesAfterCreation: 35},
				{authorEmail: "agent@helpdesk.local", body: "Internal: kemungkinan driver storage rusak setelah update. Siapkan USB recovery.", isInternal: true, minutesAfterCreation: 40},
			},
		},
		{
			title:           "Aplikasi payroll error saat ekspor PDF",
			description:     "Tombol ekspor PDF di menu payroll selalu menampilkan error 500. Ekspor Excel masih normal.",
			category:        domain.CategorySoftware,
			priority:        domain.PriorityUrgent,
			status:          domain.StatusOpen,
			requesterEmail:  "requester2@helpdesk.local",
			createdHoursAgo: 2,
			comments: []demoComment{
				{authorEmail: "requester2@helpdesk.local", body: "Ini dibutuhkan sebelum cut-off gaji besok, mohon dibantu.", minutesAfterCreation: 5},
			},
		},
		{
			title:           "Tidak bisa connect VPN kantor dari rumah",
			description:     "VPN client gagal di tahap autentikasi saat digunakan dari jaringan rumah. Dari kantor normal.",
			category:        domain.CategoryNetwork,
			priority:        domain.PriorityHigh,
			status:          domain.StatusInProgress,
			requesterEmail:  "requester3@helpdesk.local",
			assigneeEmail:   "agent2@helpdesk.local",
			createdHoursAgo: 50,
			comments: []demoComment{
				{authorEmail: "requester3@helpdesk.local", body: "Sudah coba ganti password, tetap gagal di tahap autentikasi.", minutesAfterCreation: 12},
				{authorEmail: "agent2@helpdesk.local", body: "Saya kirim ulang profil VPN baru lewat email, silakan impor dan coba lagi.", minutesAfterCreation: 90},
				{authorEmail: "agent2@helpdesk.local", body: "Internal: cek log RADIUS, ada beberapa kegagalan MFA dari IP rumah user.", isInternal: true, minutesAfterCreation: 95},
			},
		},
		{
			title:           "Reset password email SSO",
			description:     "Password email SSO terkunci karena salah memasukkan password terlalu banyak. Mohon direset.",
			category:        domain.CategoryAccess,
			priority:        domain.PriorityMedium,
			status:          domain.StatusResolved,
			requesterEmail:  "requester@helpdesk.local",
			assigneeEmail:   "agent@helpdesk.local",
			createdHoursAgo: 74,
			comments: []demoComment{
				{authorEmail: "agent@helpdesk.local", body: "Password sementara sudah dikirim ke nomor WhatsApp terdaftar. Wajib diganti saat login pertama.", minutesAfterCreation: 20},
				{authorEmail: "requester@helpdesk.local", body: "Sudah berhasil login dan ganti password. Terima kasih.", minutesAfterCreation: 45},
				{authorEmail: "agent@helpdesk.local", body: "Internal: akun sempat terkunci 30 menit karena 10 percobaan gagal.", isInternal: true, minutesAfterCreation: 22},
			},
		},
		{
			title:           "Printer lantai 3 macet kertas",
			description:     "Printer bersama di lantai 3 terus menampilkan paper jam meskipun kertas sudah dibersihkan.",
			category:        domain.CategoryHardware,
			priority:        domain.PriorityLow,
			status:          domain.StatusClosed,
			requesterEmail:  "requester2@helpdesk.local",
			assigneeEmail:   "agent@helpdesk.local",
			createdHoursAgo: 122,
			comments: []demoComment{
				{authorEmail: "agent@helpdesk.local", body: "Roller printer sudah dibersihkan dan sensor kertas direset. Silakan dicoba lagi.", minutesAfterCreation: 60},
				{authorEmail: "requester2@helpdesk.local", body: "Sudah normal, tiket bisa ditutup.", minutesAfterCreation: 120},
				{authorEmail: "agent@helpdesk.local", body: "Internal: perlu pengecekan rutin tiap bulan, roller mulai aus.", isInternal: true, minutesAfterCreation: 65},
			},
		},
		{
			title:           "Monitor berkedip saat kabel HDMI digerakkan",
			description:     "Monitor eksternal berkedip hitam setiap kabel HDMI sedikit bergerak. Sudah coba kabel lain, tetap sama.",
			category:        domain.CategoryHardware,
			priority:        domain.PriorityMedium,
			status:          domain.StatusOpen,
			requesterEmail:  "requester3@helpdesk.local",
			createdHoursAgo: 4,
		},
		{
			title:           "Request akses folder finance di server",
			description:     "Saya butuh akses read-only ke folder finance untuk menyusun laporan triwulan. Atasan sudah menyetujui.",
			category:        domain.CategoryAccess,
			priority:        domain.PriorityMedium,
			status:          domain.StatusOpen,
			requesterEmail:  "requester2@helpdesk.local",
			createdHoursAgo: 20,
			comments: []demoComment{
				{authorEmail: "requester2@helpdesk.local", body: "Surat persetujuan atasan sudah saya lampirkan di email terpisah.", minutesAfterCreation: 15},
			},
		},
		{
			title:           "Email tidak masuk dari vendor eksternal",
			description:     "Email dari vendor pengiriman tidak pernah masuk ke inbox, padahal di folder spam juga tidak ada.",
			category:        domain.CategoryNetwork,
			priority:        domain.PriorityHigh,
			status:          domain.StatusResolved,
			requesterEmail:  "requester3@helpdesk.local",
			assigneeEmail:   "agent2@helpdesk.local",
			createdHoursAgo: 96,
			comments: []demoComment{
				{authorEmail: "agent2@helpdesk.local", body: "Domain vendor ternyata masuk daftar blokir spam. Sudah kami whitelist, mohon dicek kembali.", minutesAfterCreation: 180},
				{authorEmail: "requester3@helpdesk.local", body: "Email vendor sudah masuk ke inbox. Terima kasih.", minutesAfterCreation: 240},
			},
		},
		{
			title:           "Instalasi Microsoft Office di laptop baru",
			description:     "Laptop pengganti belum memiliki Microsoft Office. Mohon dibantu instalasi beserta aktivasi lisensi.",
			category:        domain.CategorySoftware,
			priority:        domain.PriorityLow,
			status:          domain.StatusClosed,
			requesterEmail:  "requester@helpdesk.local",
			assigneeEmail:   "agent2@helpdesk.local",
			createdHoursAgo: 146,
			comments: []demoComment{
				{authorEmail: "agent2@helpdesk.local", body: "Office sudah terpasang dan lisensi aktif. Silakan dicek.", minutesAfterCreation: 90},
			},
		},
		{
			title:           "Aplikasi CRM lambat saat jam sibuk",
			description:     "CRM sangat lambat antara jam 10.00-12.00, terutama saat membuka daftar pelanggan dan laporan penjualan.",
			category:        domain.CategorySoftware,
			priority:        domain.PriorityUrgent,
			status:          domain.StatusInProgress,
			requesterEmail:  "requester2@helpdesk.local",
			assigneeEmail:   "agent@helpdesk.local",
			createdHoursAgo: 8,
			comments: []demoComment{
				{authorEmail: "requester2@helpdesk.local", body: "Hari ini lebih lambat dari kemarin, beberapa halaman butuh lebih dari 30 detik.", minutesAfterCreation: 8},
				{authorEmail: "agent@helpdesk.local", body: "Kami pantau query database dan cache aplikasi, akan diinfo jika ada update.", minutesAfterCreation: 25},
			},
		},
		{
			title:           "Kabel LAN ruang meeting rusak",
			description:     "Koneksi internet di ruang meeting putus-putus. Konektor kabel LAN terlihat patah di bagian pengunci.",
			category:        domain.CategoryNetwork,
			priority:        domain.PriorityLow,
			status:          domain.StatusOpen,
			requesterEmail:  "requester3@helpdesk.local",
			createdHoursAgo: 30,
		},
		{
			title:           "Permintaan pengadaan mouse wireless",
			description:     "Mouse yang dipakai saat ini tombol kirinya sudah tidak responsif. Mohon pengadaan satu unit mouse wireless.",
			category:        domain.CategoryOther,
			priority:        domain.PriorityLow,
			status:          domain.StatusOpen,
			requesterEmail:  "requester@helpdesk.local",
			createdHoursAgo: 18,
			comments: []demoComment{
				{authorEmail: "requester@helpdesk.local", body: "Kalau memungkinkan yang ukurannya pas untuk tangan kecil.", minutesAfterCreation: 30},
			},
		},
	}
}
