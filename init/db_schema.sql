-- Schema for the gotest article board (MySQL 8).
-- Run against the target database, e.g.: mysql -u <user> -p <db_name> < init/db_schema.sql

CREATE TABLE IF NOT EXISTS `articles` (
  `id`           INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `title`        VARCHAR(255) NOT NULL,
  `announcement` VARCHAR(255) NOT NULL,
  `full_text`    TEXT,
  PRIMARY KEY (`id`)
)
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8mb4
COMMENT = 'our articles';
