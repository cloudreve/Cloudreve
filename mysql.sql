-- phpMyAdmin SQL Dump
-- version 4.6.6
-- https://www.phpmyadmin.net/
--
-- Host: localhost
-- Generation Time: 2018-02-04 02:54:09
-- 服务器版本： 5.7.17
-- PHP Version: 7.1.2

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `lite`
--

-- --------------------------------------------------------

--
-- 表的结构 `sd_callback`
--

CREATE TABLE `sd_callback` (
  `id` int(11) NOT NULL,
  `callback_key` text NOT NULL,
  `pid` int(11) NOT NULL,
  `uid` int(11) NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 表的结构 `sd_chunks`
--

CREATE TABLE `sd_chunks` (
  `id` int(11) NOT NULL,
  `user` int(11) NOT NULL,
  `ctx` text NOT NULL,
  `time` datetime NOT NULL,
  `obj_name` text NOT NULL,
  `chunk_id` int(11) NOT NULL,
  `sum` int(11) NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 表的结构 `sd_corn`
--

CREATE TABLE `sd_corn` (
  `id` int(11) NOT NULL,
  `rank` int(11) NOT NULL,
  `name` text NOT NULL,
  `des` text NOT NULL,
  `last_excute` int(11) NOT NULL,
  `interval_s` int(11) NOT NULL,
  `enable` tinyint(1) NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

--
-- 转存表中的数据 `sd_corn`
--

INSERT INTO `sd_corn` (`id`, `rank`, `name`, `des`, `last_excute`, `interval_s`, `enable`) VALUES
(1, 2, 'delete_unseful_chunks', '删除分片上传产生的失效文件块', 0, 3600, 1),
(2, 1, 'delete_callback_data', '删除callback记录', 0, 86400, 1),
(3, 1, 'flush_aria2', '刷新离线下载状态', 0, 30, 1),
(4, 3, 'flush_onedrive_token', '刷新Onedrive Token', 0, 3000, 1);
-- --------------------------------------------------------

--
-- 表的结构 `sd_files`
--

CREATE TABLE `sd_files` (
  `id` int(11) NOT NULL,
  `orign_name` text NOT NULL,
  `pre_name` text NOT NULL,
  `upload_user` int(11) NOT NULL,
  `size` bigint(20) NOT NULL,
  `upload_date` datetime NOT NULL,
  `pic_info` text NOT NULL,
  `parent_folder` int(11) NOT NULL,
  `policy_id` int(11) NOT NULL,
  `dir` text NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 表的结构 `sd_folders`
--

CREATE TABLE `sd_folders` (
  `id` int(11) NOT NULL,
  `folder_name` text NOT NULL,
  `parent_folder` int(11) NOT NULL,
  `position` text NOT NULL,
  `owner` text NOT NULL,
  `date` datetime NOT NULL,
  `position_absolute` text NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

--
-- 转存表中的数据 `sd_folders`
--

INSERT INTO `sd_folders` (`id`, `folder_name`, `parent_folder`, `position`, `owner`, `date`, `position_absolute`) VALUES
(1, '根目录', 0, '.', '1', '2018-01-30 10:13:34', '/');

-- --------------------------------------------------------

--
-- 表的结构 `sd_groups`
--

CREATE TABLE `sd_groups` (
  `id` int(11) NOT NULL,
  `group_name` text NOT NULL,
  `policy_name` int(11) NOT NULL,
  `max_storage` bigint(20) NOT NULL,
  `grade_policy` text NOT NULL,
  `speed` text NOT NULL,
  `allow_share` tinyint(1) NOT NULL,
  `color` text NOT NULL,
  `policy_list` text NOT NULL,
  `range_transfer` tinyint(1) NOT NULL,
  `webdav` tinyint(1) NOT NULL,
  `aria2` text NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

--
-- 转存表中的数据 `sd_groups`
--

INSERT INTO `sd_groups` (`id`, `group_name`, `policy_name`, `max_storage`, `grade_policy`, `speed`, `allow_share`, `color`, `policy_list`, `range_transfer`, `webdav`,`aria2`) VALUES
(1, '管理员', 1, 1073741824, '', '', 1, 'danger', '1', 1, 1, "0,0,0"),
(2, '游客', 1, 0, '', '', 1, 'default', '1', 0, 0, "0,0,0"),
(3, '注册会员', 1, 52428800, '', '', 1, 'default', '1', 1, 1, "0,0,0");

-- --------------------------------------------------------

--
-- 表的结构 `sd_options`
--

CREATE TABLE `sd_options` (
  `id` int(11) NOT NULL,
  `option_name` text NOT NULL,
  `option_value` text NOT NULL,
  `option_type` text NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

--
-- 转存表中的数据 `sd_options`
--

INSERT INTO `sd_options` (`id`, `option_name`, `option_value`, `option_type`) VALUES
(1, 'siteURL', 'https://cloudreve.org/', 'basic'),
(2, 'siteName', 'Cloudreve', 'basic'),
(3, 'siteStatus', 'open', 'basic'),
(4, 'regStatus', '0', 'register'),
(5, 'defaultGroup', '3', 'register'),
(6, 'siteKeywords', '网盘，网盘', 'basic'),
(7, 'siteDes', 'Cloudreve', 'basic'),
(8, 'siteTitle', '平步云端', 'basic'),
(9, 'fromName', 'Cloudreve', 'mail'),
(10, 'fromAdress', 'no-reply@acg.blue', 'mail'),
(11, 'smtpHost', 'smtp.mxhichina.com', 'mail'),
(12, 'smtpPort', '25', 'mail'),
(13, 'replyTo', 'abslant@126.com', 'mail'),
(14, 'smtpUser', 'no-reply@acg.blue', 'mail'),
(15, 'smtpPass', '', 'mail'),
(16, 'encriptionType', 'no', 'mail'),
(22, 'maxEditSize', '100000', 'file_edit'),
(48, 'timeout', '3600', 'oss'),
(23, 'allowdVisitorDownload', 'false', 'share'),
(24, 'login_captcha', '0', 'login'),
(28, 'reg_captcha', '0', 'login'),
(29, 'email_active', '0', 'register'),
(30, 'mail_activation_template', '     <!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Transitional//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd\">\r\n<html xmlns=\"http://www.w3.org/1999/xhtml\" style=\"font-family: \'Helvetica Neue\', Helvetica, Arial, sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\">\r\n<head>\r\n<meta name=\"viewport\" content=\"width=device-width\" />\r\n<meta http-equiv=\"Content-Type\" content=\"text/html; charset=UTF-8\" />\r\n<title>容量超额提醒</title>\r\n\r\n\r\n<style type=\"text/css\">\r\nimg {\r\nmax-width: 100%;\r\n}\r\nbody {\r\n-webkit-font-smoothing: antialiased; -webkit-text-size-adjust: none; width: 100% !important; height: 100%; line-height: 1.6em;\r\n}\r\nbody {\r\nbackground-color: #f6f6f6;\r\n}\r\n@media only screen and (max-width: 640px) {\r\n  body {\r\n    padding: 0 !important;\r\n  }\r\n  h1 {\r\n    font-weight: 800 !important; margin: 20px 0 5px !important;\r\n  }\r\n  h2 {\r\n    font-weight: 800 !important; margin: 20px 0 5px !important;\r\n  }\r\n  h3 {\r\n    font-weight: 800 !important; margin: 20px 0 5px !important;\r\n  }\r\n  h4 {\r\n    font-weight: 800 !important; margin: 20px 0 5px !important;\r\n  }\r\n  h1 {\r\n    font-size: 22px !important;\r\n  }\r\n  h2 {\r\n    font-size: 18px !important;\r\n  }\r\n  h3 {\r\n    font-size: 16px !important;\r\n  }\r\n  .container {\r\n    padding: 0 !important; width: 100% !important;\r\n  }\r\n  .content {\r\n    padding: 0 !important;\r\n  }\r\n  .content-wrap {\r\n    padding: 10px !important;\r\n  }\r\n  .invoice {\r\n    width: 100% !important;\r\n  }\r\n}\r\n</style>\r\n</head>\r\n\r\n<body itemscope itemtype=\"http://schema.org/EmailMessage\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; -webkit-font-smoothing: antialiased; -webkit-text-size-adjust: none; width: 100% !important; height: 100%; line-height: 1.6em; background-color: #f6f6f6; margin: 0;\" bgcolor=\"#f6f6f6\">\r\n\r\n<table class=\"body-wrap\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; width: 100%; background-color: #f6f6f6; margin: 0;\" bgcolor=\"#f6f6f6\"><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0;\" valign=\"top\"></td>\r\n		<td class=\"container\" width=\"600\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; display: block !important; max-width: 600px !important; clear: both !important; margin: 0 auto;\" valign=\"top\">\r\n			<div class=\"content\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; max-width: 600px; display: block; margin: 0 auto; padding: 20px;\">\r\n				<table class=\"main\" width=\"100%\" cellpadding=\"0\" cellspacing=\"0\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; border-radius: 3px; background-color: #fff; margin: 0; border: 1px solid #e9e9e9;\" bgcolor=\"#fff\"><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"alert alert-warning\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 16px; vertical-align: top; color: #fff; font-weight: 500; text-align: center; border-radius: 3px 3px 0 0; background-color: #009688; margin: 0; padding: 20px;\" align=\"center\" bgcolor=\"#FF9F00\" valign=\"top\">\r\n							激活{siteTitle}账户\r\n						</td>\r\n					</tr><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"content-wrap\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 20px;\" valign=\"top\">\r\n							<table width=\"100%\" cellpadding=\"0\" cellspacing=\"0\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"content-block\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;\" valign=\"top\">\r\n										亲爱的 <strong style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\">{userName}</strong> ：\r\n									</td>\r\n								</tr><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"content-block\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;\" valign=\"top\">\r\n										感谢您注册{siteTitle},请点击下方按钮完成账户激活。\r\n									</td>\r\n								</tr><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"content-block\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;\" valign=\"top\">\r\n										<a href=\"{activationUrl}\" class=\"btn-primary\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; color: #FFF; text-decoration: none; line-height: 2em; font-weight: bold; text-align: center; cursor: pointer; display: inline-block; border-radius: 5px; text-transform: capitalize; background-color: #009688; margin: 0; border-color: #009688; border-style: solid; border-width: 10px 20px;\">激活账户</a>\r\n									</td>\r\n								</tr><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"content-block\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;\" valign=\"top\">\r\n										感谢您选择{siteTitle}。\r\n									</td>\r\n								</tr></table></td>\r\n					</tr></table><div class=\"footer\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; width: 100%; clear: both; color: #999; margin: 0; padding: 20px;\">\r\n					<table width=\"100%\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"aligncenter content-block\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 12px; vertical-align: top; color: #999; text-align: center; margin: 0; padding: 0 0 20px;\" align=\"center\" valign=\"top\">此邮件由系统自动发送，请不要直接回复。</td>\r\n						</tr></table></div></div>\r\n		</td>\r\n		<td style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0;\" valign=\"top\"></td>\r\n	</tr></table></body>\r\n</html>\r\n', 'mail_template'),
(31, 'forget_captcha', '0', 'login'),
(32, 'mail_reset_pwd_template', '     <!DOCTYPE html PUBLIC \"-//W3C//DTD XHTML 1.0 Transitional//EN\" \"http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd\">\r\n<html xmlns=\"http://www.w3.org/1999/xhtml\" style=\"font-family: \'Helvetica Neue\', Helvetica, Arial, sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\">\r\n<head>\r\n<meta name=\"viewport\" content=\"width=device-width\" />\r\n<meta http-equiv=\"Content-Type\" content=\"text/html; charset=UTF-8\" />\r\n<title>重设密码</title>\r\n\r\n\r\n<style type=\"text/css\">\r\nimg {\r\nmax-width: 100%;\r\n}\r\nbody {\r\n-webkit-font-smoothing: antialiased; -webkit-text-size-adjust: none; width: 100% !important; height: 100%; line-height: 1.6em;\r\n}\r\nbody {\r\nbackground-color: #f6f6f6;\r\n}\r\n@media only screen and (max-width: 640px) {\r\n  body {\r\n    padding: 0 !important;\r\n  }\r\n  h1 {\r\n    font-weight: 800 !important; margin: 20px 0 5px !important;\r\n  }\r\n  h2 {\r\n    font-weight: 800 !important; margin: 20px 0 5px !important;\r\n  }\r\n  h3 {\r\n    font-weight: 800 !important; margin: 20px 0 5px !important;\r\n  }\r\n  h4 {\r\n    font-weight: 800 !important; margin: 20px 0 5px !important;\r\n  }\r\n  h1 {\r\n    font-size: 22px !important;\r\n  }\r\n  h2 {\r\n    font-size: 18px !important;\r\n  }\r\n  h3 {\r\n    font-size: 16px !important;\r\n  }\r\n  .container {\r\n    padding: 0 !important; width: 100% !important;\r\n  }\r\n  .content {\r\n    padding: 0 !important;\r\n  }\r\n  .content-wrap {\r\n    padding: 10px !important;\r\n  }\r\n  .invoice {\r\n    width: 100% !important;\r\n  }\r\n}\r\n</style>\r\n</head>\r\n\r\n<body itemscope itemtype=\"http://schema.org/EmailMessage\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; -webkit-font-smoothing: antialiased; -webkit-text-size-adjust: none; width: 100% !important; height: 100%; line-height: 1.6em; background-color: #f6f6f6; margin: 0;\" bgcolor=\"#f6f6f6\">\r\n\r\n<table class=\"body-wrap\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; width: 100%; background-color: #f6f6f6; margin: 0;\" bgcolor=\"#f6f6f6\"><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0;\" valign=\"top\"></td>\r\n		<td class=\"container\" width=\"600\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; display: block !important; max-width: 600px !important; clear: both !important; margin: 0 auto;\" valign=\"top\">\r\n			<div class=\"content\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; max-width: 600px; display: block; margin: 0 auto; padding: 20px;\">\r\n				<table class=\"main\" width=\"100%\" cellpadding=\"0\" cellspacing=\"0\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; border-radius: 3px; background-color: #fff; margin: 0; border: 1px solid #e9e9e9;\" bgcolor=\"#fff\"><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"alert alert-warning\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 16px; vertical-align: top; color: #fff; font-weight: 500; text-align: center; border-radius: 3px 3px 0 0; background-color: #2196F3; margin: 0; padding: 20px;\" align=\"center\" bgcolor=\"#FF9F00\" valign=\"top\">\r\n							重设{siteTitle}密码\r\n						</td>\r\n					</tr><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"content-wrap\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 20px;\" valign=\"top\">\r\n							<table width=\"100%\" cellpadding=\"0\" cellspacing=\"0\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"content-block\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;\" valign=\"top\">\r\n										亲爱的 <strong style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\">{userName}</strong> ：\r\n									</td>\r\n								</tr><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"content-block\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;\" valign=\"top\">\r\n										请点击下方按钮完成密码重设。如果非你本人操作，请忽略此邮件。\r\n									</td>\r\n								</tr><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"content-block\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;\" valign=\"top\">\r\n										<a href=\"{resetUrl}\" class=\"btn-primary\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; color: #FFF; text-decoration: none; line-height: 2em; font-weight: bold; text-align: center; cursor: pointer; display: inline-block; border-radius: 5px; text-transform: capitalize; background-color: #2196F3; margin: 0; border-color: #2196F3; border-style: solid; border-width: 10px 20px;\">重设密码</a>\r\n									</td>\r\n								</tr><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"content-block\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0; padding: 0 0 20px;\" valign=\"top\">\r\n										感谢您选择{siteTitle}。\r\n									</td>\r\n								</tr></table></td>\r\n					</tr></table><div class=\"footer\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; width: 100%; clear: both; color: #999; margin: 0; padding: 20px;\">\r\n					<table width=\"100%\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><tr style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; margin: 0;\"><td class=\"aligncenter content-block\" style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 12px; vertical-align: top; color: #999; text-align: center; margin: 0; padding: 0 0 20px;\" align=\"center\" valign=\"top\">此邮件由系统自动发送，请不要直接回复。</td>\r\n						</tr></table></div></div>\r\n		</td>\r\n		<td style=\"font-family: \'Helvetica Neue\',Helvetica,Arial,sans-serif; box-sizing: border-box; font-size: 14px; vertical-align: top; margin: 0;\" valign=\"top\"></td>\r\n	</tr></table></body>\r\n</html>\r\n', 'mail_template'),
(49, 'database_version', '3', 'version'),
(43, 'hot_share_num', '10', 'share'),
(44, 'gravatar_server', 'https://v2ex.assets.uxengine.net/gravatar/', 'avatar'),
(45, 'admin_color_body', 'fixed-nav sticky-footer bg-light', 'admin'),
(46, 'admin_color_nav', 'navbar navbar-expand-lg fixed-top navbar-light bg-light', 'admin'),
(47, 'js_code', '<script type=\"text/javascript\">\r\n\r\n</script>', 'basic'),
(50, 'sendfile', '0', 'download'),
(51, 'header', 'X-Sendfile', 'download'),
(52, 'aria2_tmppath', '/path/to/public/download', 'aria2'),
(53, 'aria2_token', 'your token', 'aria2'),
(54, 'aria2_rpcurl', 'http://127.0.0.1:6800/', 'aria2'),
(55, 'aria2_options', '{\"max-tries\":5}', 'aria2'),
(56, 'task_queue_token', '', 'task');
-- --------------------------------------------------------

--
-- 表的结构 `sd_policy`
--

CREATE TABLE `sd_policy` (
  `id` int(11) NOT NULL,
  `policy_name` text NOT NULL,
  `policy_type` text NOT NULL,
  `server` text NOT NULL,
  `bucketname` text NOT NULL,
  `bucket_private` tinyint(1) NOT NULL,
  `url` text NOT NULL,
  `ak` text NOT NULL,
  `sk` text NOT NULL,
  `op_name` text NOT NULL,
  `op_pwd` text NOT NULL,
  `filetype` text NOT NULL,
  `mimetype` text NOT NULL,
  `max_size` bigint(20) NOT NULL,
  `autoname` tinyint(1) NOT NULL,
  `dirrule` text NOT NULL,
  `namerule` text NOT NULL,
  `origin_link` tinyint(1) NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

--
-- 转存表中的数据 `sd_policy`
--

INSERT INTO `sd_policy` (`id`, `policy_name`, `policy_type`, `server`, `bucketname`, `bucket_private`, `url`, `ak`, `sk`, `op_name`, `op_pwd`, `filetype`, `mimetype`, `max_size`, `autoname`, `dirrule`, `namerule`, `origin_link`) VALUES
(1, '默认上传策略', 'local', '/Upload', '0', 0, 'http://cloudreve.org/public/uploads/', '0', '0', '0', '0', '[]', '0', 10485760, 1, '{date}/{uid}', '{uid}_{randomkey8}_{originname}', 0);

-- --------------------------------------------------------

--
-- 表的结构 `sd_shares`
--

CREATE TABLE `sd_shares` (
  `id` int(11) NOT NULL,
  `type` text NOT NULL,
  `share_time` datetime NOT NULL,
  `owner` int(11) NOT NULL,
  `source_name` text NOT NULL,
  `origin_name` text NOT NULL,
  `download_num` int(11) NOT NULL,
  `view_num` int(11) NOT NULL,
  `source_type` text NOT NULL,
  `share_key` text NOT NULL,
  `share_pwd` text NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 表的结构 `sd_storage_pack`
--

CREATE TABLE `sd_storage_pack` (
  `id` int(11) NOT NULL,
  `p_name` text NOT NULL,
  `uid` int(11) NOT NULL,
  `act_time` bigint(20) NOT NULL,
  `dlay_time` bigint(20) NOT NULL,
  `pack_size` bigint(11) NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

-- --------------------------------------------------------

--
-- 表的结构 `sd_users`
--

CREATE TABLE `sd_users` (
  `id` int(11) NOT NULL,
  `user_email` varchar(100) NOT NULL,
  `user_nick` varchar(50) NOT NULL,
  `user_pass` varchar(255) NOT NULL,
  `user_date` timestamp NOT NULL,
  `user_status` int(11) NOT NULL,
  `user_group` int(11) NOT NULL,
  `group_primary` int(11) NOT NULL,
  `user_activation_key` varchar(255) NOT NULL,
  `used_storage` bigint(20) NOT NULL,
  `two_step` text NOT NULL,
  `delay_time` bigint(20) NOT NULL,
  `avatar` text NOT NULL,
  `profile` tinyint(1) NOT NULL,
  `webdav_key` text NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8;

--
-- 转存表中的数据 `sd_users`
--

INSERT INTO `sd_users` (`id`, `user_email`, `user_nick`, `user_pass`, `user_date`, `user_status`, `user_group`, `group_primary`, `user_activation_key`, `used_storage`, `two_step`, `delay_time`, `avatar`, `profile`, `webdav_key`) VALUES
(1, 'admin@cloudreve.org', 'Admin', 'd8446059f8846a2c111a7f53515665fb', '2018-01-30 02:13:34', 0, 1, 0, 'n', 0, '0', 0, 'default', 1, 'd8446059f8846a2c111a7f53515665fb');
CREATE TABLE `sd_download` (
  `id` int(11) NOT NULL,
  `pid` text NOT NULL,
  `path_id` text NOT NULL,
  `owner` int(11) NOT NULL,
  `save_dir` text NOT NULL,
  `status` text NOT NULL,
  `last_update` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `msg` text NOT NULL,
  `info` text NOT NULL,
  `source` text NOT NULL,
  `file_index` int(11) NOT NULL,
  `is_single` tinyint(1) NOT NULL,
  `total_size` bigint(20) NOT NULL
) ENGINE=MyISAM DEFAULT CHARSET=utf8mb4;
ALTER TABLE `sd_download`
  ADD PRIMARY KEY (`id`);
ALTER TABLE `sd_download`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
--
-- Indexes for dumped tables
--
--
-- 表的结构 `sd_task`
--

CREATE TABLE `sd_task` (
  `id` int(11) NOT NULL,
  `task_name` text NOT NULL,
  `attr` text NOT NULL,
  `type` text NOT NULL,
  `status` text NOT NULL,
  `uid` int(11) NOT NULL,
  `addtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

--
-- Indexes for table `sd_callback`
--
ALTER TABLE `sd_callback`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `sd_chunks`
--
ALTER TABLE `sd_chunks`
  ADD PRIMARY KEY (`id`);
ALTER TABLE `sd_task`
  ADD PRIMARY KEY (`id`);
--
-- Indexes for table `sd_corn`
--
ALTER TABLE `sd_corn`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `sd_files`
--
ALTER TABLE `sd_files`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `sd_folders`
--
ALTER TABLE `sd_folders`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `sd_groups`
--
ALTER TABLE `sd_groups`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `sd_options`
--
ALTER TABLE `sd_options`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `sd_policy`
--
ALTER TABLE `sd_policy`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `sd_shares`
--
ALTER TABLE `sd_shares`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `sd_storage_pack`
--
ALTER TABLE `sd_storage_pack`
  ADD PRIMARY KEY (`id`);

--
-- Indexes for table `sd_users`
--
ALTER TABLE `sd_users`
  ADD PRIMARY KEY (`id`);

--
-- 在导出的表使用AUTO_INCREMENT
--

--
-- 使用表AUTO_INCREMENT `sd_callback`
--
ALTER TABLE `sd_callback`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
--
-- 使用表AUTO_INCREMENT `sd_chunks`
--
ALTER TABLE `sd_chunks`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
--
-- 使用表AUTO_INCREMENT `sd_corn`
--
ALTER TABLE `sd_corn`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=6;
--
-- 使用表AUTO_INCREMENT `sd_files`
--
ALTER TABLE `sd_files`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
--
-- 使用表AUTO_INCREMENT `sd_folders`
--
ALTER TABLE `sd_folders`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;
ALTER TABLE `sd_task`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
--
-- 使用表AUTO_INCREMENT `sd_groups`
--
ALTER TABLE `sd_groups`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=4;
--
-- 使用表AUTO_INCREMENT `sd_options`
--
ALTER TABLE `sd_options`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=50;
--
-- 使用表AUTO_INCREMENT `sd_policy`
--
ALTER TABLE `sd_policy`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;
--
-- 使用表AUTO_INCREMENT `sd_shares`
--
ALTER TABLE `sd_shares`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
--
-- 使用表AUTO_INCREMENT `sd_storage_pack`
--
ALTER TABLE `sd_storage_pack`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT;
--
-- 使用表AUTO_INCREMENT `sd_users`
--
ALTER TABLE `sd_users`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT, AUTO_INCREMENT=2;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
