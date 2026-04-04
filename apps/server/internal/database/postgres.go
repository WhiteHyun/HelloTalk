package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // ← 이 import가 핵심

	"github.com/whitehyun/HelloTalk/apps/server/internal/config"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	// 1. DSN(Data Source Name) 조립
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		cfg.DBSSLMode,
	)

	// 2. sql.Open
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 3. 커넥션 풀 설정
	db.SetMaxOpenConns(25)                 // DB에 동시에 열 수 있는 최대 연결 수. 이 이상의 요청은 연결이 반환될 때까지 대기
	db.SetMaxIdleConns(5)                  // 사용 안 하는 상태로 유지해둘 연결 수. 다음 요청이 오면 새로 연결할 필요 없이 재사용
	db.SetConnMaxLifetime(5 * time.Minute) // 연결 하나의 최대 수명. 이 시간이 지나면 닫고 새로 만든다. DB 쪽 타임아웃이나 로드밸런서 대응용

	// 4. Ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Successfully connected to database")
	return db, nil
}
