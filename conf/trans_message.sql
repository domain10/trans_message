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
  `ip_arr` text CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT 'IP地址多个用英文逗号隔开，用于白名单过滤',
  `status` tinyint NOT NULL DEFAULT '1' COMMENT '状态：1启用 0停用',
  `query_url` varchar(128) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '查询确认消息的地址',
  `query_times` tinyint unsigned NOT NULL DEFAULT '10' COMMENT '最大查询次数',
  `query_interval` smallint unsigned NOT NULL DEFAULT '10' COMMENT '查询间隔时长（秒）',
  `notify_url` varchar(128) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '接收消息通知的地址',
  `notify_times` tinyint unsigned NOT NULL DEFAULT '10' COMMENT '最大通知次数',
  `notify_interval` smallint unsigned NOT NULL DEFAULT '10' COMMENT '通知间隔时长（秒）',
  `describe` varchar(256) NOT NULL DEFAULT '' COMMENT '描述',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `u_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=10 DEFAULT CHARSET=utf8;

/*Table structure for table `message_list` */

DROP TABLE IF EXISTS `message_list`;

CREATE TABLE `message_list` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '自增',
  `mid` bigint unsigned NOT NULL COMMENT '消息id，唯一',
  `app_name` varchar(32) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT '来自哪个应用',
  `from_address` varchar(128) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL COMMENT '来源ip',
  `message_type` tinyint NOT NULL DEFAULT '1' COMMENT '消息类型：1同时执行的消息 2依次执行的消息',
  `list` json NOT NULL COMMENT '消息列表 [{"to_app":"接收应用","to_url":"接收地址，没有该参数时则从注册应用里获取","content":"消息内容","status":"1成功 其它失败"}]',
  `status` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '0新消息 1未confirm 2已确认通知中 3取消 4失败 5成功 6人工处理成功',
  `describe` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '' COMMENT '消息状态描述',
  `query_count` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '查询确认次数',
  `notify_count` tinyint unsigned NOT NULL DEFAULT '0' COMMENT '消息通知次数',
  `alarm_count` int NOT NULL DEFAULT '0' COMMENT '失败告警次数',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uqe_mid` (`mid`),
  KEY `c_index` (`create_time`)
) ENGINE=InnoDB AUTO_INCREMENT=48 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
