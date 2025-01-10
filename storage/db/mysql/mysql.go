package mysql

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"h2blog/pkg/config"
	"h2blog/pkg/logger"
	"h2blog/storage/db/dbLogger"
)

// ConnectMysql 函数用于连接 MySQL 数据库
func ConnectMysql() *gorm.DB {
	// 记录日志，表示准备连接 MySQL 数据库
	logger.Info("准备连接 MySQL 数据库")
	// 构建 DSN (Data Source Name) 字符串，用于连接 MySQL 数据库
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Asia%%2FShanghai",
		config.MySQLConfig.User,     // MySQL 用户名
		config.MySQLConfig.Password, // MySQL 密码
		config.MySQLConfig.Host,     // MySQL 服务器地址
		config.MySQLConfig.Port,     // MySQL 服务器端口
		config.MySQLConfig.DB,       // 数据库名称
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
			Logger: dbLogger.NewDbLogger(),
		},
	)

	// 检查数据库连接是否成功
	if err != nil {
		// 如果连接失败，记录错误日志并退出程序
		msg := fmt.Sprintf("MySQL 数据库连接失败: %v", err)
		logger.Panic(msg)
	}

	// 创建相关基础数据表
	if !tableExists(db, "LOGIN_RECORD") {
		// 创建 LOGIN_RECORD 表
		err = db.Exec(createLoginRecordTable).Error
		if err != nil {
			handleError("创建 LOGIN_RECORD 表失败", err)
		}
	}
	if !tableExists(db, "H2_BLOG_INFO") {
		// 创建 H2_BLOG_INFO 表
		err = db.Exec(createH2BlogInfoTableSQL).Error
		if err != nil {
			handleError("创建 H2_BLOG_INFO 表失败", err)
		}
	}
	if !tableExists(db, "H2_IMG_INFO") {
		err = db.Exec(createH2ImgInfoTableSQL).Error
		if err != nil {
			handleError("创建 H2_IMG_INFO 表失败", err)
		}
	}

	// 记录日志，表示 MySQL 数据库连接成功
	logger.Info("MySQL 数据库连接成功")

	return db
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
