package sqlite

import (
	"context"
	"path/filepath"
	"sparrow_blog_server/pkg/config"
	"sparrow_blog_server/pkg/filetool"
	"sparrow_blog_server/pkg/logger"
	"sparrow_blog_server/storage/db/dblogger"
	"sparrow_blog_server/storage/db/sqlscript"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectSqlite(ctx context.Context) (*gorm.DB, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	// 生成数据目录
	dataDir := filepath.Dir(config.Sqlite.Path)
	if err := filetool.EnsureDir(dataDir); err != nil {
		handleError("数据目录操作失败", err)
	}

	db, err := gorm.Open(
		sqlite.Open(config.Sqlite.Path),
		&gorm.Config{
			Logger: dblogger.NewDbLogger(),
		},
	)

	// 检查数据库连接是否成功
	if err != nil {
		// 如果连接失败，记录错误日志并退出程序
		handleError("Sqlite 数据库连接失败", err)
	}

	// 创建相关基础数据表
	if !tableExists(db, "BLOG") {
		err = db.Exec(sqlscript.CreateBlogTableSQL).Error
		if err != nil {
			handleError("创建 BLOG 表失败", err)
		}
	}

	if !tableExists(db, "BLOG_READ_COUNT") {
		err = db.Exec(sqlscript.CreateBlogReadCountTableSQL).Error
		if err != nil {
			handleError("创建 BLOG_READ_COUNT 表失败", err)
		}
	}

	if !tableExists(db, "CATEGORY") {
		err = db.Exec(sqlscript.CreateCategoryTableSQL).Error
		if err != nil {
			handleError("创建 CATEGORY 表失败", err)
		}
	}

	if !tableExists(db, "TAG") {
		err = db.Exec(sqlscript.CreateTagTableSQL).Error
		if err != nil {
			handleError("创建 TAG 表失败", err)
		}
	}

	if !tableExists(db, "BLOG_TAG") {
		err = db.Exec(sqlscript.CreateBlogTagTableSQL).Error
		if err != nil {
			handleError("创建 BLOG_TAG 表失败", err)
		}
	}

	if !tableExists(db, "IMG") {
		err = db.Exec(sqlscript.CreateImgTableSQL).Error
		if err != nil {
			handleError("创建 IMG 表失败", err)
		}
	}

	if !tableExists(db, "COMMENT") {
		err = db.Exec(sqlscript.CreateCommentTableSQL).Error
		if err != nil {
			handleError("创建 COMMENT 表失败", err)
		}
	}

	logger.Info("Sqlite 数据库连接成功")

	return db, nil
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
