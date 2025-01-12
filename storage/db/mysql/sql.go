package mysql

// 创建一个名为 LOGIN_RECORD 的表，如果该表不存在则创建
const createLoginRecordTable = `
	CREATE TABLE IF NOT EXISTS LOGIN_RECORD
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
	    title      		VARCHAR(50) 				NOT NULL 														COMMENT '博客标题',
	    brief      		VARCHAR(255)				NOT NULL 														COMMENT '博客简介',
	    create_time		TIMESTAMP					NOT NULL DEFAULT CURRENT_TIMESTAMP 								COMMENT '创建时间',
	    update_time		TIMESTAMP					NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP 	COMMENT '更新时间',
	    CONSTRAINT uc_title UNIQUE (title) COMMENT '标题唯一约束'
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
	    create_time 	TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP 								COMMENT '创建时间',
	    update_time 	TIMESTAMP 					NOT NULL	DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP 	COMMENT '更新时间',
	    CONSTRAINT uc_title UNIQUE (img_name) COMMENT '名称唯一约束'
	) COMMENT='图片信息表' 
	  ENGINE = InnoDB
	  DEFAULT CHARSET = utf8mb4
	  COLLATE = utf8mb4_unicode_ci;
`
