package mysql

// 创建表
const createDatabaseSql = `CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`

// 创建一个名为 BLOG_INFO 的表，如果该表不存在则创建
const createBlogTableSQL = `
	CREATE TABLE IF NOT EXISTS BLOG
	(
	    blog_id         	VARCHAR(16)      	PRIMARY KEY NOT NULL  											COMMENT '博客ID',
	    blog_title    	 	VARCHAR(50)      	NOT NULL UNIQUE       											COMMENT '博客标题',
	    blog_image_id		VARCHAR(16)			NOT NULL														COMMENT '博客图片 ID',
	    blog_brief    	 	VARCHAR(255)     	NOT NULL              											COMMENT '博客简介',
	    category_id     	VARCHAR(16)      	NOT NULL              											COMMENT '分类ID（逻辑外键）',
	    blog_state        	TINYINT(1)       	NOT NULL              											COMMENT '博客状态（0-禁用 1-启用）',
	    blog_words_num  	SMALLINT UNSIGNED 	NOT NULL             									 		COMMENT '博客字数',
	    blog_is_top     	TINYINT(1)       	NOT NULL              											COMMENT '是否置顶（0-否 1-是）',
	    create_time     	TIMESTAMP        	NOT NULL DEFAULT CURRENT_TIMESTAMP 								COMMENT '创建时间',
	    update_time     	TIMESTAMP        	NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP 	COMMENT '更新时间',
	    INDEX (blog_id),
	    INDEX (blog_title),
	    INDEX (create_time),
	    INDEX (category_id) -- 为逻辑外键添加索引
	) COMMENT = '博客信息表'
	  ENGINE = InnoDB
	  DEFAULT CHARSET = utf8mb4
	  COLLATE = utf8mb4_unicode_ci;
`

const createCategoryTableSQL = `
	CREATE TABLE IF NOT EXISTS CATEGORY
	(
	    category_id   	VARCHAR(16)  PRIMARY KEY NOT NULL 											COMMENT '分类ID',
	    category_name 	VARCHAR(50)  NOT NULL UNIQUE 												COMMENT '分类名称',
	    create_time   	TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP 							COMMENT '创建时间',
	    update_time   	TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
	    INDEX (category_name)
	) ENGINE = InnoDB
	  DEFAULT CHARSET = utf8mb4
	  COLLATE = utf8mb4_unicode_ci;
`

const createTagTableSQL = `
	CREATE TABLE IF NOT EXISTS TAG
	(
	    tag_id     	VARCHAR(16)  PRIMARY KEY NOT NULL 											COMMENT '标签ID',
	    tag_name  	VARCHAR(50)  NOT NULL UNIQUE 												COMMENT '标签名称',
	    create_time TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP 							COMMENT '创建时间',
	    update_time TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
	    INDEX (tag_name)
	) ENGINE = InnoDB
	  DEFAULT CHARSET = utf8mb4
	  COLLATE = utf8mb4_unicode_ci;
`

// 多对多关联表
const createBlogTagTableSQL = `
	CREATE TABLE IF NOT EXISTS BLOG_TAG
	(
	    blog_id 	VARCHAR(16) NOT NULL COMMENT '博客ID',
	    tag_id 		VARCHAR(16) NOT NULL COMMENT '标签ID',
	    PRIMARY KEY (blog_id, tag_id) -- 联合主键
	) ENGINE = InnoDB
	  DEFAULT CHARSET = utf8mb4
	  COLLATE = utf8mb4_unicode_ci;
`

const createImgTableSQL = `
	CREATE TABLE IMG
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
	CREATE TABLE IF NOT EXISTS COMMENT
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

const createFriendLinkTableSQL = `
	CREATE TABLE IF NOT EXISTS FRIEND_LINK
	(
		friend_link_id 			VARCHAR(16) 	PRIMARY KEY NOT NULL 															COMMENT '友链 ID',
		friend_link_name 		VARCHAR(50) 				NOT NULL 															COMMENT '友链名称',
		friend_link_url 		VARCHAR(200) 				NOT NULL 	UNIQUE													COMMENT '友链 URL',
		friend_link_avatar_url 	VARCHAR(200) 																					COMMENT '友链头像 URL',
		friend_describe			VARCHAR(500)																					COMMENT '友链描述',
		display					bool						NOT NULL 	DEFAULT FALSE											COMMENT '是否展示',
		create_time 			TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP 								COMMENT '创建时间',
		update_time 			TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP 	COMMENT '更新时间',
		INDEX (friend_link_id),
		INDEX (friend_link_name),
		INDEX (create_time)
	) COMMENT = '友链表'
	  ENGINE = InnoDB
	  DEFAULT CHARSET = utf8mb4
	  COLLATE = utf8mb4_unicode_ci;
`
