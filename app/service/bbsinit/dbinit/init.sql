CREATE TABLE `users`
(
    `id`                  bigint unsigned NOT NULL AUTO_INCREMENT,
    `username`            varchar(255)    NOT NULL DEFAULT '',
    `nickname`            varchar(255)    NOT NULL DEFAULT '',
    `email`               varchar(255)    NOT NULL DEFAULT '',
    `password`            varchar(255)    NOT NULL DEFAULT '',
    `mobile_area_code`    varchar(16),
    `mobile_phone_number` varchar(64),
    `status`              tinyint(4)      NOT NULL DEFAULT '0' COMMENT '状态：0正常 1冻结 ',
    `validate`            tinyint(4)      NOT NULL DEFAULT '0' COMMENT '是否验证通过: 0未通过/未验证 1 验证通过',
    `prestige`            bigint          NOT NULL default 0 COMMENT '声望',
    `created_at`          datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`          datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`          datetime                 DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_users_username` (`username`),
    UNIQUE KEY `idx_users_email` (`email`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '用户';

alter table users
    add `mobile_area_code`    varchar(16) after password,
    add `mobile_phone_number` varchar(64) after mobile_area_code,
    add `status`              tinyint(4) NOT NULL DEFAULT '0' COMMENT '状态：0正常 1冻结 ' after mobile_phone_number,
    add `validate`            tinyint(4) NOT NULL DEFAULT '0' COMMENT '是否验证通过: 0未通过/未验证 1 验证通过' after status,
    add `prestige`            bigint     NOT NULL default 0 COMMENT '声望' after validate;

alter table users
    add `nickname` varchar(255) NOT NULL DEFAULT '' after username;


CREATE TABLE `user_follow`
(
    `id`             bigint unsigned NOT NULL AUTO_INCREMENT,
    `user_id`        bigint unsigned NOT NULL,
    `follow_user_id` bigint unsigned NOT NULL,
    `created_at`     datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '用户关注表';

CREATE TABLE user_points
(
    user_id        bigint unsigned NOT NULL AUTO_INCREMENT,
    current_points bigint          not null default 0,
    `created_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '用户积分表';


CREATE TABLE points_record
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
    user_id       bigint unsigned NOT NULL default 0,
    change_reason varchar(255),
    `created_at`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_user_id` (`user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '积分记录表';

CREATE TABLE article_tag
(
    id           bigint unsigned NOT NULL AUTO_INCREMENT,
    tag          VARCHAR(255)    not null default '',
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '标签';

CREATE TABLE article_category
(
    id           bigint unsigned NOT NULL AUTO_INCREMENT,
    category     VARCHAR(255)    not null default '',
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '分类';

CREATE TABLE article_category_rs
(
    id                  bigint unsigned NOT NULL AUTO_INCREMENT,
    article_id          bigint unsigned NOT NULL,
    article_category_id bigint unsigned NOT NULL,
    `effective`         int             not null default 0,
    `created_at`        datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`        datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    unique key uiq_category_article (article_category_id, article_id)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '文章分类关系表，目的是为了让某些交叉文章可以更好的广播出去';

CREATE TABLE `articles`
(
    `id`             bigint unsigned NOT NULL AUTO_INCREMENT,
    `title`          varchar(512)    NOT NULL DEFAULT '',
    `content`        text,
    `user_id`        bigint unsigned NOT NULL DEFAULT '0',
    `article_status` tinyint(4)      NOT NULL DEFAULT '0' COMMENT '文章状态：0 草稿 1 发布',
    `process_status` tinyint(4)      NOT NULL DEFAULT '0' COMMENT '管理状态：0 正常 1 封禁',
    `created_at`     datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`     datetime                 DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '文章';

alter table articles
    modify `user_id` bigint unsigned NOT NULL DEFAULT '0';

alter table articles
    add `article_status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '文章状态：0 草稿 1 发布' after user_id,
    add `process_status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '管理状态：0 正常 1 封禁' after article_status;
alter table articles
    add `type` tinyint(4) NOT NULL DEFAULT '0' COMMENT '文章类型：0 博文，1教程，2问答，3分享' after `content`;

CREATE TABLE `comment`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT,
    `article_id` bigint          NOT NULL DEFAULT '0',
    `content`    text,
    `user_id`    bigint          NOT NULL DEFAULT '0',
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '评论';

alter table comment add reply_id   bigint          NOT NULL DEFAULT '0' after content;

CREATE TABLE `reply`
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT,
    `article_id` bigint unsigned NOT NULL DEFAULT '0',
    `user_id`    bigint unsigned NOT NULL DEFAULT '0',
    `target_id`  bigint unsigned NOT NULL DEFAULT '0' comment '目标id',
    `content`    text,
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '回复表';

alter table reply
    modify `article_id` bigint unsigned NOT NULL DEFAULT '0',
    modify `user_id` bigint unsigned NOT NULL DEFAULT '0',
    modify `target_id` bigint unsigned NOT NULL DEFAULT '0' comment '目标id';


CREATE TABLE event_notification
(
    `id`                  bigint unsigned NOT NULL AUTO_INCREMENT,
    `user_id`             VARCHAR(255),
    received_notification TEXT,
    event_type            VARCHAR(50),
    `created_at`          datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`          datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '事件通知表';


CREATE TABLE user_role_rs
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT,
    `user_id`    bigint unsigned NOT NULL default 0,
    `role_id`    bigint unsigned NOT NULL default 0,
    `effective`  int             not null default 0,
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '用户角色表表';

CREATE TABLE role
(
    `id`         bigint unsigned NOT NULL AUTO_INCREMENT,
    `role_name`  VARCHAR(255),
    `effective`  int             not null default 0,
    `created_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at` datetime                 DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '角色表';

CREATE TABLE role_permission_rs
(
    `id`            bigint unsigned NOT NULL AUTO_INCREMENT,
    `role_id`       bigint unsigned NOT NULL default 0,
    `permission_id` bigint unsigned NOT NULL default 0,
    `effective`     int             not null default 0,
    `created_at`    datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`    datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `deleted_at`    datetime                 DEFAULT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '角色权限表';

CREATE TABLE opt_record
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT,
    `opt_user_id` bigint unsigned NOT NULL default 0,
    `opt_type`    int             not null default 0,
    `target_type` int             not null default 0,
    `target_id`   varchar(255)    not null default '',
    `opt_info`    text,
    `created_at`  datetime        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_0900_ai_ci COMMENT '操作记录表';
