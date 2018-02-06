<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006~2016 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: liu21st <liu21st@gmail.com>
// +----------------------------------------------------------------------
use think\Route;
Route::rule([
	'Upload/mkblk/:chunkSize'=>'index/Upload/chunk',
	'Upload/mkfile/:fileSize/key/:keyValue/fname/:fname/path/:path'=>'index/Upload/mkFile',
	's/:key'=>'index/Share/index',
	'Share/Download/:key'=>'index/Share/Download',
	'Share/Preview/:key'=>'index/Share/Preview',
	'Share/ListFile/:key'=>'index/Share/ListFile',
	'Login'=>'index/Member/LoginForm',
	'Member/emailActivate/:key'=>'index/Member/emailActivate',
	'Member/resetPwd/:key'=>'index/Member/resetPwd',
	'Callback/Payment/Jinshajiang' => 'index/Callback/Jinshajiang',
	'Explore/S/:key' => 'index/Explore/S',
	'Member/Avatar/:uid/:size' => ['Member/Avatar',[],['uid'=>'\d+']],
	'Profile/:uid' => ['Profile/index',[],['uid'=>'\d+']],
	'Callback/Payment/Jinshajiang' => 'index/Callback/Jinshajiang',
]);