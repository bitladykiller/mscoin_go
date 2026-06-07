// Package mysqlx 集中管理所有服务的 MySQL 数据库连接初始化。
//
// 本包封装了 github.com/jmoiron/sqlx 库，提供：
//   - 统一的数据库连接创建方式
//   - 自动驼峰到蛇形命名转换（CamelCase -> snake_case）
//   - 兼容传统 GORM 标签的字段映射
//   - 连接池配置
//
// 命名转换说明：
//   - 结构体字段 `UserName` 自动映射到数据库列 `user_name`
//   - 支持 `gorm:"column:custom_name"` 标签自定义列名
//   - 这使得遗留的 GORM 结构体无需修改即可使用 sqlx
//
// 使用场景：
//   - 各微服务的数据库连接初始化
//   - 与 sqlx 的 Query/Select/Exec 方法配合使用
package mysqlx

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	// MySQL 驱动，匿名导入以注册驱动
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

// mapperOnce 确保全局 NameMapper 只设置一次。
// 使用 sync.Once 避免多次设置导致的竞态条件。
var mapperOnce sync.Once

// Config 捕获服务间共享的 MySQL 连接设置。
//
// 字段说明：
//   - DataSource: MySQL 数据源名称（DSN），格式为 "user:password@tcp(host:port)/dbname?options"
//   - MaxOpenConns: 最大打开连接数，建议设置为与数据库服务器的 max_connections 匹配
//   - MaxIdleConns: 最大空闲连接数，建议设置为 MaxOpenConns 的一部分（如 1/4）
//   - ConnMaxLifetime: 连接最大存活时间（秒），建议设置为 30 分钟以避免长时间连接问题
//   - ConnMaxIdleTime: 空闲连接最大存活时间（秒），建议设置为 5 分钟
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
//
// 参数：
//   - cfg: MySQL 配置
//
// 返回值：
//   - *sqlx.DB: 数据库连接实例
//   - error: 连接失败时返回错误
//
// 使用示例：
//
//	db, err := New(Config{
//	    DataSource:      "root:password@tcp(127.0.0.1:3306)/mscoin?parseTime=true",
//	    MaxOpenConns:    100,
//	    MaxIdleConns:    25,
//	    ConnMaxLifetime: 1800,
//	})
//	if err != nil {
//	    // 处理错误
//	}
//	defer db.Close()
func New(cfg Config) (*sqlx.DB, error) {
	// 打开数据库连接
	db, err := sqlx.Open("mysql", cfg.DataSource)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	// 设置全局字段名映射器（驼峰转蛇形）
	// 使用 sync.Once 确保只设置一次
	mapperOnce.Do(func() {
		sqlx.NameMapper = toSnake
	})

	// 设置当前连接的映射器，支持 GORM 标签
	// 这允许结构体使用 `gorm:"column:custom_name"` 自定义列名
	db.Mapper = reflectx.NewMapperTagFunc("gorm", toSnake, parseGormColumn)

	// 配置连接池
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	// 验证连接是否可用
	// 这确保服务启动时快速失败，而不是在第一次查询时才发现连接问题
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}

// parseGormColumn 解析 GORM 标签中的 column 指令。
//
// GORM 标签格式示例：`gorm:"column:user_name;type:varchar(100)"`
// 此函数提取其中的 column 值用于字段映射。
//
// 参数：
//   - tag: GORM 标签字符串
//
// 返回值：
//   - string: 列名，如果未找到 column 指令则返回原始标签
func parseGormColumn(tag string) string {
	// 按分号分割标签的各个部分
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		// 查找 column: 前缀
		if strings.HasPrefix(part, "column:") {
			return strings.TrimPrefix(part, "column:")
		}
	}
	// 未找到 column 指令，返回原始标签
	return tag
}

// toSnake 将驼峰命名转换为蛇形命名。
//
// 转换规则：
//   - 大写字母前插入下划线（除了第一个字符）
//   - 大写字母转换为小写
//
// 示例：
//   - "UserName" -> "user_name"
//   - "ID" -> "id"
//   - "HTTPServer" -> "h_t_t_p_server"
//
// 参数：
//   - value: 驼峰命名字符串
//
// 返回值：
//   - string: 蛇形命名字符串
func toSnake(value string) string {
	var builder strings.Builder
	// 预分配空间，蛇形命名最多比原字符串长 1/4
	builder.Grow(len(value) + 4)

	for i, r := range value {
		if unicode.IsUpper(r) {
			// 大写字母前插入下划线（除了第一个字符）
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