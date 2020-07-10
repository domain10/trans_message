/*
SQLyog Ultimate v12.08 (64 bit)
MySQL - 8.0.20 : Database - trans_message
*********************************************************************
*/

/*!40101 SET NAMES utf8 */;

/*!40101 SET SQL_MODE=''*/;

/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;
CREATE DATABASE /*!32312 IF NOT EXISTS*/`trans_message` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;

USE `trans_message`;

/*Table structure for table `application` */

DROP TABLE IF EXISTS `application`;

CREATE TABLE `application` (
  `id` mediumint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `name` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '应用名称',
  `app_key` char(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '应用的key',
  `ip` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT 'IP地址，用于内部白名单过滤',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：1启用 2停用',
  `query_url` varchar(128) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '查询确认消息的地址',
  `query_times` tinyint unsigned NOT NULL DEFAULT '10' COMMENT '最大查询次数',
  `notify_url` varchar(128) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '消息通知回调的地址',
  `notify_times` tinyint unsigned NOT NULL DEFAULT '10' COMMENT '最多通知次数',
  `describe` varchar(256) NOT NULL DEFAULT '' COMMENT '描述',
  `create_time` int unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `update_time` int unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uqe_key` (`app_key`)
) ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8;

/*Table structure for table `message_list` */

DROP TABLE IF EXISTS `message_list`;

CREATE TABLE `message_list` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `mid` bigint unsigned NOT NULL COMMENT '消息id',
  `app_key` char(32) NOT NULL DEFAULT '' COMMENT '应用的key',
  `from_address` varchar(128) NOT NULL COMMENT '来源',
  `message_type` tinyint NOT NULL DEFAULT '1' COMMENT '消息类型：1同时执行的消息 2依次执行的消息',
  `list` text CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '消息列表',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态：1新消息 2通知中 3成功 4失败 5转人工处理',
  `describe` varchar(256) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '消息状态描述',
  `notify_count` smallint unsigned NOT NULL DEFAULT '0' COMMENT '消息通知次数',
  `create_time` int unsigned NOT NULL COMMENT '创建时间',
  `update_time` int unsigned NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uqe_mid` (`mid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

/*Table structure for table `message_list_20190601` */

DROP TABLE IF EXISTS `message_list_20190601`;

CREATE TABLE `message_list_20190601` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `mid` bigint unsigned NOT NULL COMMENT '消息id',
  `app_key` char(32) NOT NULL DEFAULT '' COMMENT '应用的key',
  `from_address` varchar(128) NOT NULL COMMENT '来源',
  `message_type` tinyint NOT NULL DEFAULT '1' COMMENT '消息类型：1同时执行的消息 2依次执行的消息',
  `list` text CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '消息列表',
  `status` tinyint unsigned NOT NULL DEFAULT '1' COMMENT '状态：1新消息 2通知中 3成功 4失败 5转人工处理',
  `describe` varchar(256) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '消息状态描述',
  `notify_count` smallint unsigned NOT NULL DEFAULT '0' COMMENT '消息通知次数',
  `create_time` int unsigned NOT NULL COMMENT '创建时间',
  `update_time` int unsigned NOT NULL COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uqe_mid` (`mid`)
) ENGINE=InnoDB AUTO_INCREMENT=41 DEFAULT CHARSET=utf8;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
