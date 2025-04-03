package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"h2blog_server/pkg/config"
	"h2blog_server/pkg/logger"
	"h2blog_server/storage/db/dblogger"
)

// createLoginRecordTableSQL 是用于创建库
func createDatabase(dbName string) error {
	// 连接 MySQL（不指定库名）
	db, err := sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=true&loc=Asia%%2FShanghai",
			config.MySQL.User,     // MySQL 用户名
			config.MySQL.Password, // MySQL 密码
			config.MySQL.Host,     // MySQL 服务器地址
			config.MySQL.Port,     // MySQL 服务器端口
		),
	)
	if err != nil {
		return err
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error(err.Error())
		}
	}(db)

	// 创建数据库
	_, err = db.Exec(fmt.Sprintf(createDatabaseSql, dbName))
	return err
}

// ConnectMysql 函数用于连接 MySQL 数据库
func ConnectMysql(ctx context.Context) (*gorm.DB, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// 创建数据库
	logger.Info("准备创建数据库")
	err := createDatabase(config.MySQL.DB)
	if err != nil {
		logger.Panic("数据库创建失败: %v", err)
	}

	// 记录日志，表示准备连接 MySQL 数据库
	logger.Info("准备连接 MySQL 数据库")
	// 构建 DSN (Data Source Name) 字符串，用于连接 MySQL 数据库
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Asia%%2FShanghai",
		config.MySQL.User,     // MySQL 用户名
		config.MySQL.Password, // MySQL 密码
		config.MySQL.Host,     // MySQL 服务器地址
		config.MySQL.Port,     // MySQL 服务器端口
		config.MySQL.DB,       // 数据库名称
	)

	// 使用 GORM 打开 MySQL 数据库连接
	db, err := gorm.Open(
		mysql.New(
			mysql.Config{
				DSN:                       dsn,   // 数据源名称
				DefaultStringSize:         256,   // 默认字符串大小
				DisableDatetimePrecision:  false, // 是否禁用时间精度
				DontSupportRenameIndex:    true,  // 是否不支持重命名索引
				DontSupportRenameColumn:   true,  // 是否不支持重命名列
				SkipInitializeWithVersion: false, // 是否跳过版本初始化
			},
		),
		&gorm.Config{
			Logger: dblogger.NewDbLogger(),
		},
	)

	// 检查数据库连接是否成功
	if err != nil {
		// 如果连接失败，记录错误日志并退出程序
		msg := fmt.Sprintf("MySQL 数据库连接失败: %v", err)
		logger.Panic(msg)
	}

	// 创建相关基础数据表
	if !tableExists(db, "H2_LOGIN_RECORD") {
		// 创建 H2_LOGIN_RECORD 表
		err = db.Exec(createLoginRecordTable).Error
		if err != nil {
			handleError("创建 H2_LOGIN_RECORD 表失败", err)
		}
	}
	if !tableExists(db, "H2_BLOG") {
		// 创建 H2_BLOG 表
		err = db.Exec(createH2BlogTableSQL).Error
		if err != nil {
			handleError("创建 H2_BLOG 表失败", err)
		}
	}
	if !tableExists(db, "H2_CATEGORY") {
		err = db.Exec(createH2CategoryTableSQL).Error
		if err != nil {
			handleError("创建 H2_CATEGORY 表失败", err)
		}
	}
	if !tableExists(db, "H2_TAG") {
		err = db.Exec(createH2TagTableSQL).Error
		if err != nil {
			handleError("创建 H2_TAG 表失败", err)
		}
	}
	if !tableExists(db, "H2_BLOG_TAG") {
		err = db.Exec(createH2BlogTagTableSQL).Error
		if err != nil {
			handleError("创建 H2_BLOG_TAG 表失败", err)
		}
	}
	if !tableExists(db, "H2_IMG") {
		err = db.Exec(createH2ImgInfoTableSQL).Error
		if err != nil {
			handleError("创建 H2_IMG 表失败", err)
		}
	}
	if !tableExists(db, "H2_COMMENT") {
		err = db.Exec(createH2CommentTableSQL).Error
		if err != nil {
			handleError("创建 H2_COMMENT 表失败", err)
		}
	}
	if !tableExists(db, "H2_FRIEND_LINK") {
		err = db.Exec(createH2FriendLinkTableSQL).Error
		if err != nil {
			handleError("创建 H2_FRIEND_LINK 表失败", err)
		}
	}

	// 记录日志，表示 MySQL 数据库连接成功
	logger.Info("MySQL 数据库连接成功")

	return db, err
}

func tableExists(db *gorm.DB, tableName string) bool {
	var count int
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = ?", tableName).Scan(&count)
	return count > 0
}

func handleError(msg string, err error) {
	logger.Error(msg + ": " + err.Error())
	panic(msg + ": " + err.Error())
}
