package sqlscript

const CreateBlogTableSQL = `
	CREATE TABLE IF NOT EXISTS BLOG
	(
	    blog_id         	VARCHAR(16)      	PRIMARY KEY NOT NULL,  					-- 博客ID
	    blog_title    	 	VARCHAR(50)      	NOT NULL UNIQUE,       					-- 博客标题
	    blog_image_id		VARCHAR(16)			NOT NULL,								-- 博客图片 ID
	    blog_brief    	 	VARCHAR(255)     	NOT NULL,              					-- 博客简介
	    category_id     	VARCHAR(16)      	NOT NULL,              					-- 分类ID（逻辑外键）
	    blog_state        	INTEGER       		NOT NULL,              					-- 博客状态（0-禁用 1-启用）
	    blog_words_num  	INTEGER 			NOT NULL,             					-- 博客字数
	    blog_is_top     	INTEGER       		NOT NULL,              					-- 是否置顶（0-否 1-是）
	    create_time     	TIMESTAMP        	NOT NULL DEFAULT CURRENT_TIMESTAMP, 	-- 创建时间
	    update_time     	TIMESTAMP        	NOT NULL DEFAULT CURRENT_TIMESTAMP 		-- 更新时间
	); -- 博客信息表
`

const CreateBlogReadCountTableSQL = `
	CREATE TABLE IF NOT EXISTS BLOG_READ_COUNT
	(
		read_id				VARCHAR(16)      	PRIMARY KEY NOT NULL, 	-- 阅读记录 ID，由博客 ID 与日期生成
		blog_id				VARCHAR(16)			NOT NULL,
		read_count 			INT					NOT NULL DEFAULT 0,
		read_date			CHAR(8)				NOT NULL DEFAULT '' 	-- 阅读日期
	); -- 博客阅读量表
`

const CreateCategoryTableSQL = `
	CREATE TABLE IF NOT EXISTS CATEGORY
	(
	    category_id   	VARCHAR(16)  PRIMARY KEY NOT NULL, 										-- 分类ID
	    category_name 	VARCHAR(50)  NOT NULL UNIQUE, 											-- 分类名称
	    create_time 		TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP, 	-- 创建时间
		update_time 		TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP 	-- 更新时间
	);
`

const CreateTagTableSQL = `
	CREATE TABLE IF NOT EXISTS TAG
	(
	    tag_id     	VARCHAR(16)  PRIMARY KEY NOT NULL, 										
	    tag_name  	VARCHAR(50)  NOT NULL UNIQUE, 											
	    create_time 		TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP, 
		update_time 		TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP 
	);
`

const CreateBlogTagTableSQL = `
	CREATE TABLE IF NOT EXISTS BLOG_TAG
	(
	    blog_id 	VARCHAR(16) NOT NULL, 	-- 博客ID
	    tag_id 		VARCHAR(16) NOT NULL, 	-- 标签ID
	    PRIMARY KEY (blog_id, tag_id) 		-- 联合主键
	);
`

const CreateImgTableSQL = `
	CREATE TABLE IF NOT EXISTS IMG
	(
	    img_id 			VARCHAR(16) 	PRIMARY KEY NOT NULL,								-- 图片ID
	    img_name 		VARCHAR(255) 				NOT NULL	UNIQUE, 					-- 图片名称
	    img_type 		VARCHAR(10) 				NOT NULL, 								-- 图片类型
	    create_time 	TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP, 	-- 创建时间
	    update_time 	TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP 	-- 更新时间
	); -- 图片信息表
`

const CreateCommentTableSQL = `
	CREATE TABLE IF NOT EXISTS COMMENT
	(
		comment_id 			VARCHAR(16) 	PRIMARY KEY NOT NULL, 								-- 评论 ID
		commenter_email 	VARCHAR(50)  				NOT NULL,  								-- 评论者邮箱
		blog_id 			VARCHAR(16),				        								-- 博客 ID
		original_poster_id 	VARCHAR(16), 				 										-- 楼主评论 ID
		reply_to_comment_id VARCHAR(16), 				 										-- 回复的评论 ID
		comment_content 	TEXT 						NOT NULL, 								-- 评论内容(最大支持64KB)
		create_time 		TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP, 	-- 创建时间
		update_time 		TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP 	-- 更新时间
	); -- 评论主表
`
