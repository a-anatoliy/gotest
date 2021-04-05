USE `test_lucky` ;

-- -----------------------------------------------------
-- Table `test_lucky`.`articles`
-- -----------------------------------------------------
DROP TABLE IF EXISTS `test_lucky`.`articles` ;

CREATE TABLE IF NOT EXISTS `test_lucky`.`articles` (
  `id` TINYINT(1) UNSIGNED NOT NULL AUTO_INCREMENT,
  `announcement` VARCHAR(255) NOT NULL,
  `title` VARCHAR(255) NOT NULL,
  `full_text` TEXT,
  PRIMARY KEY (`id`))
ENGINE = InnoDB
DEFAULT CHARACTER SET = utf8
COMMENT = 'our articles';

