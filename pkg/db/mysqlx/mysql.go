// Package mysqlx 集中管理所有服务的 MySQL 初始化。
package mysqlx

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

var mapperOnce sync.Once

// Config 捕获服务间共享的 MySQL 连接设置。
type Config struct {
	DataSource      string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int64
	ConnMaxIdleTime int64
}

// New 打开 MySQL 连接并标准化结构体字段映射。
//
// 重构保留了通过 `sqlx` 使用显式 SQL 的方式，但也保留了便捷的字段映射，
// 以便遗留结构体继续携带历史的 `gorm:"column:"` 标签。
func New(cfg Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("mysql", cfg.DataSource)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	mapperOnce.Do(func() {
		sqlx.NameMapper = toSnake
	})
	db.Mapper = reflectx.NewMapperTagFunc("gorm", toSnake, parseGormColumn)

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}

func parseGormColumn(tag string) string {
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	return tag
}

func toSnake(value string) string {
	var builder strings.Builder
	builder.Grow(len(value) + 4)
	for i, r := range value {
		if unicode.IsUpper(r) {
			if i > 0 {
				builder.WriteByte('_')
			}
			builder.WriteRune(unicode.ToLower(r))
			continue
		}
		builder.WriteRune(r)
	}
	return builder.String()
}
