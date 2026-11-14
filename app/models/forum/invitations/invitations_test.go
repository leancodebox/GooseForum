package invitations

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/leancodebox/GooseForum/app/bundles/connect/dbconnect"
	"github.com/leancodebox/GooseForum/app/models/forum/taskQueue"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestCreateAndRedeemHonorsUseLimit(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate invitations: %v", err)
	}
	expires := time.Now().Add(time.Hour)
	entity, raw, err := CreateWithDB(db, 710001, 2, &expires)
	if err != nil {
		t.Fatalf("CreateWithDB() err=%v", err)
	}
	if entity.Hash != Digest(raw) || entity.Prefix != raw[:16] {
		t.Fatalf("stored secret metadata does not match generated code")
	}
	if _, err := RedeemWithDB(db, raw, 710002, time.Now()); err != nil {
		t.Fatalf("first redeem: %v", err)
	}
	used, err := RedeemWithDB(db, raw, 710003, time.Now())
	if err != nil || used.UsedCount != 2 || used.LastUsedUserId != 710003 {
		t.Fatalf("second redeem = %+v, err=%v; want count 2 and last user", used, err)
	}
	if _, err := RedeemWithDB(db, raw, 710004, time.Now()); !errors.Is(err, ErrExhausted) {
		t.Fatalf("third redeem err=%v, want ErrExhausted", err)
	}
}

func TestPageIncludesInvitationEmailTaskStatus(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&Entity{}, &taskQueue.Entity{}); err != nil {
		t.Fatalf("migrate invitation delivery tables: %v", err)
	}
	creator := uint64(780001)
	db.Where("created_by = ?", creator).Delete(&Entity{})
	t.Cleanup(func() { db.Where("created_by = ?", creator).Delete(&Entity{}) })
	entity, _, err := CreateWithDB(db, creator, 1, nil)
	if err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	status := uint8(taskQueue.StatusFailed)
	task := &taskQueue.Entity{Type: "registration_invitation", Status: status, TaskJson: `{}`}
	if err := db.Create(task).Error; err != nil {
		t.Fatalf("create email task: %v", err)
	}
	t.Cleanup(func() { db.Delete(&taskQueue.Entity{}, task.Id) })
	if err := db.Model(&Entity{}).Where("id = ?", entity.Id).Update("email_task_id", task.Id).Error; err != nil {
		t.Fatalf("link email task: %v", err)
	}
	rows, _, _, total, err := Page(PageQuery{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("Page() error: %v", err)
	}
	if total < 1 {
		t.Fatalf("Page() total = %d, want invitation", total)
	}
	for _, row := range rows {
		if row.Id == entity.Id {
			if row.EmailTaskStatus == nil || *row.EmailTaskStatus != status {
				t.Fatalf("email task status = %v, want %d", row.EmailTaskStatus, status)
			}
			return
		}
	}
	t.Fatalf("created invitation %d not found in page", entity.Id)
}

func TestPageQueryGeneratesPortableJoinAcrossDialects(t *testing.T) {
	tests := []struct {
		name      string
		dialector gorm.Dialector
	}{
		{name: "sqlite", dialector: sqlite.Open(":memory:")},
		{name: "mysql", dialector: mysql.New(mysql.Config{DSN: "root@tcp(localhost:3306)/gooseforum", SkipInitializeWithVersion: true})},
		{name: "postgres", dialector: postgres.New(postgres.Config{DSN: "host=localhost user=postgres dbname=gooseforum", PreferSimpleProtocol: true})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := gorm.Open(tt.dialector, &gorm.Config{DryRun: true, DisableAutomaticPing: true})
			if err != nil {
				t.Fatalf("open dry-run database: %v", err)
			}
			statement := pageQuery(db).Order("created_at DESC").Limit(20).Find(&[]Entity{}).Statement
			sql := statement.SQL.String()
			if !strings.Contains(strings.ToUpper(sql), "LEFT JOIN") || !strings.Contains(sql, "task_queue") || !strings.Contains(sql, "email_task_status") {
				t.Fatalf("generated SQL = %q; want task status left join", sql)
			}
		})
	}
}

func TestPostgresReservationAllowsOnlyOneWinner(t *testing.T) {
	dsn := os.Getenv("GOOSEFORUM_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("GOOSEFORUM_POSTGRES_TEST_DSN is not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("underlying PostgreSQL connection: %v", err)
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(4)
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate invitations: %v", err)
	}
	assertReservationAllowsOnlyOneWinner(t, db, 740101, false)
	t.Cleanup(func() { _ = sqlDB.Close() })
}

func TestMySQLReservationAllowsOnlyOneWinner(t *testing.T) {
	dsn := os.Getenv("GOOSEFORUM_MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("GOOSEFORUM_MYSQL_TEST_DSN is not set")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open MySQL: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("underlying MySQL connection: %v", err)
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(4)
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate invitations: %v", err)
	}
	assertReservationAllowsOnlyOneWinner(t, db, 740102, false)
	t.Cleanup(func() { _ = sqlDB.Close() })
}

func TestReservationAllowsOnlyOneWinnerAcrossConnections(t *testing.T) {
	path := t.TempDir() + "/invitations.db"
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", path)), &gorm.Config{})
	if err != nil {
		t.Fatalf("open multi-connection sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("underlying sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(4)
	sqlDB.SetMaxIdleConns(4)
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate invitations: %v", err)
	}
	assertReservationAllowsOnlyOneWinner(t, db, 740001, true)
}

func assertReservationAllowsOnlyOneWinner(t *testing.T, db *gorm.DB, creator uint64, retryLocked bool) {
	t.Helper()
	db.Where("created_by = ?", creator).Delete(&Entity{})
	t.Cleanup(func() { db.Where("created_by = ?", creator).Delete(&Entity{}) })
	_, raw, err := CreateWithDB(db, creator, 1, nil)
	if err != nil {
		t.Fatalf("create invitation: %v", err)
	}

	var wg sync.WaitGroup
	results := make(chan error, 4)
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var reserveErr error
			for attempt := 0; attempt < 5; attempt++ {
				_, _, reserveErr = ReserveWithDB(db, raw, time.Now(), time.Minute)
				if reserveErr == nil || !retryLocked || !strings.Contains(reserveErr.Error(), "locked") {
					break
				}
				time.Sleep(20 * time.Millisecond)
			}
			results <- reserveErr
		}()
	}
	wg.Wait()
	close(results)

	wins := 0
	for reserveErr := range results {
		if reserveErr == nil {
			wins++
		} else if !errors.Is(reserveErr, ErrReserved) {
			t.Fatalf("concurrent reserve err=%v, want ErrReserved for losers", reserveErr)
		}
	}
	if wins != 1 {
		t.Fatalf("concurrent reserve winners=%d, want exactly one", wins)
	}
}

func TestRedeemRejectsExpiredAndRevokedInvitations(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate invitations: %v", err)
	}
	expires := time.Now().Add(time.Minute)
	_, expiredRaw, err := CreateWithDB(db, 720001, 1, &expires)
	if err != nil {
		t.Fatalf("create expired: %v", err)
	}
	if _, err := RedeemWithDB(db, expiredRaw, 720002, expires.Add(time.Second)); !errors.Is(err, ErrExpired) {
		t.Fatalf("expired redeem err=%v, want ErrExpired", err)
	}
	entity, revokedRaw, err := CreateWithDB(db, 720003, 1, nil)
	if err != nil {
		t.Fatalf("create revocable: %v", err)
	}
	now := time.Now()
	if err := db.Model(&Entity{}).Where("id = ?", entity.Id).Update("revoked_at", now).Error; err != nil {
		t.Fatalf("revoke invitation: %v", err)
	}
	if _, err := RedeemWithDB(db, revokedRaw, 720004, now); !errors.Is(err, ErrRevoked) {
		t.Fatalf("revoked redeem err=%v, want ErrRevoked", err)
	}
}

func TestReservationCanCommitOrReleaseWithoutLeakingAUse(t *testing.T) {
	db := dbconnect.Connect()
	if err := db.AutoMigrate(&Entity{}); err != nil {
		t.Fatalf("migrate invitations: %v", err)
	}
	now := time.Now()
	_, raw, err := CreateWithDB(db, 730001, 1, nil)
	if err != nil {
		t.Fatalf("create invitation: %v", err)
	}
	reserved, reservation, err := ReserveWithDB(db, raw, now, time.Minute)
	if err != nil {
		t.Fatalf("reserve invitation: %v", err)
	}
	if _, _, err := ReserveWithDB(db, raw, now, time.Minute); !errors.Is(err, ErrReserved) {
		t.Fatalf("second reserve err=%v, want ErrReserved", err)
	}
	if reserved.ReservationHash != Digest(reservation) {
		t.Fatalf("reservation hash was not recorded")
	}
	if err := ReleaseWithDB(db, raw, "wrong", now); err != nil {
		t.Fatalf("release wrong reservation: %v", err)
	}
	if _, _, err := ReserveWithDB(db, raw, now, time.Minute); !errors.Is(err, ErrReserved) {
		t.Fatalf("wrong release should retain reservation, err=%v", err)
	}
	if _, err := CommitWithDB(db, raw, "gf_res_invalid", 730002, now); !errors.Is(err, ErrReserved) {
		t.Fatalf("invalid commit err=%v, want ErrReserved", err)
	}
	if committed, err := CommitWithDB(db, raw, reservation, 730002, now); err != nil || committed.UsedCount != 1 {
		t.Fatalf("commit = %+v, err=%v; want one use", committed, err)
	}
}
