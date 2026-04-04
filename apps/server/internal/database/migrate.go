package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func Migrate(db *sql.DB, migrationsDir string) error {
	// schema_migrations 테이블이 없으면 생성
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// 이미 적용된 최대 버전 조회
	var currentVersion int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&currentVersion)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	// migrations 디렉토리에서 .up.sql 파일 목록 읽기
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	type migration struct {
		version  int
		filename string
	}

	var migrations []migration
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		// 파일명에서 버전 번호 추출: "001_create_users.up.sql" → 1
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}

		if version > currentVersion {
			migrations = append(migrations, migration{version: version, filename: name})
		}
	}

	// 버전 순으로 정렬
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	// 각 마이그레이션을 트랜잭션으로 실행
	for _, m := range migrations {
		sqlBytes, err := os.ReadFile(filepath.Join(migrationsDir, m.filename))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", m.filename, err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(string(sqlBytes)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration %s: %w", m.filename, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", m.version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", m.filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", m.filename, err)
		}

		log.Printf("Applied migration: %s", m.filename)
	}

	return nil
}
