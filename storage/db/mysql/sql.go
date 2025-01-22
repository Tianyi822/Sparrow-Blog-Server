package mysql

// 创建表
const createDatabaseSql = `CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`

// 创建一个名为 LOGIN_RECORD 的表，如果该表不存在则创建
const createLoginRecordTable = `
	CREATE TABLE IF NOT EXISTS H2_LOGIN_RECORD
	(
		id          INT 		 PRIMARY KEY AUTO_INCREMENT NOT NULL       					     	COMMENT '记录ID',
		ipv4        VARCHAR(15)                             NOT NULL						     	COMMENT 'IPv4地址',
		ipv6        VARCHAR(39)                             NOT NULL						     	COMMENT 'IPv6地址',
		login_time  TIMESTAMP 								NOT NULL 	DEFAULT CURRENT_TIMESTAMP	COMMENT '登录时间',
		logout_time TIMESTAMP 								  	  		DEFAULT NULL             	COMMENT '登出时间',
		INDEX (ipv4),
		INDEX (ipv6),
		INDEX (login_time)
	) COMMENT = '登录记录表'
	  ENGINE = InnoDB
	  DEFAULT CHARSET = utf8mb4
	  COLLATE = utf8mb4_unicode_ci;
`

// 创建一个名为 H2_BLOG_INFO 的表，如果该表不存在则创建
const createH2BlogInfoTableSQL = `
	CREATE TABLE H2_BLOG_INFO
	(
	    blog_id    		VARCHAR(16) 	PRIMARY KEY NOT NULL 														COMMENT '博客ID',
	    title      		VARCHAR(50) 				NOT NULL UNIQUE													COMMENT '博客标题',
	    brief      		VARCHAR(255)				NOT NULL 														COMMENT '博客简介',
	    create_time		TIMESTAMP					NOT NULL DEFAULT CURRENT_TIMESTAMP 								COMMENT '创建时间',
	    update_time		TIMESTAMP					NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP 	COMMENT '更新时间',
	    INDEX (blog_id),
	    INDEX (title),
	    INDEX (create_time)
	) COMMENT = '博客信息表'
	  ENGINE = InnoDB
	  DEFAULT CHARSET = utf8mb4
	  COLLATE = utf8mb4_unicode_ci;
`

const createH2ImgInfoTableSQL = `
	CREATE TABLE H2_IMG_INFO 
	(
	    img_id 			VARCHAR(16) 	PRIMARY KEY NOT NULL															COMMENT '图片ID',
	    img_name 		VARCHAR(255) 				NOT NULL	UNIQUE 													COMMENT '图片名称',
	    img_type 		VARCHAR(10) 				NOT NULL 															COMMENT '图片类型',
	    create_time 	TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP 								COMMENT '创建时间',
	    update_time 	TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP 	COMMENT '更新时间',
	    INDEX (img_id),
	    INDEX (img_name),
	    INDEX (create_time)
	) COMMENT='图片信息表' 
	  ENGINE = InnoDB
	  DEFAULT CHARSET = utf8mb4
	  COLLATE = utf8mb4_unicode_ci;
`

const createCommentTableSQL = `
	CREATE TABLE IF NOT EXISTS H2_COMMENT
	(
		comment_id 			VARCHAR(16) 	PRIMARY KEY NOT NULL 															COMMENT '评论ID',
		user_name     		VARCHAR(50)     			NOT NULL 															COMMENT '用户名',
		user_email 			VARCHAR(50)  				NOT NULL  															COMMENT '用户邮箱',
		user_url        	VARCHAR(200) 				NOT NULL 															COMMENT '用户邮箱',
		blog_id 			VARCHAR(16) 				NOT NULL 															COMMENT '博客ID',
		original_poster_id 	VARCHAR(16) 				NOT NULL 															COMMENT '楼主评论ID',
		content 			TEXT 						NOT NULL 															COMMENT '评论内容(最大支持64KB)',
		create_time 		TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP 								COMMENT '创建时间',
		update_time 		TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP 	COMMENT '更新时间',
		INDEX (comment_id),
		INDEX (original_poster_id),
		INDEX (create_time)
	) COMMENT = '评论主表'
	  ENGINE = InnoDB
	  DEFAULT CHARSET = utf8mb4
	  COLLATE = utf8mb4_unicode_ci;
`
