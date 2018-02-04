<?php
// +----------------------------------------------------------------------
// | ThinkPHP [ WE CAN DO IT JUST THINK ]
// +----------------------------------------------------------------------
// | Copyright (c) 2006-2016 http://thinkphp.cn All rights reserved.
// +----------------------------------------------------------------------
// | Licensed ( http://www.apache.org/licenses/LICENSE-2.0 )
// +----------------------------------------------------------------------
// | Author: 流年 <liu21st@gmail.com>
// +----------------------------------------------------------------------

// 应用公共文件
function array_column_fix($array,$key){
	if(count($array) == count($array,1)){
		return array(0 =>$array[$key]);
	}else{
		return array_column($array,$key);
	}
}
function getAllowedExt($ext){
	$returnValue = "";
	foreach (json_decode($ext,true) as $key => $value) {
		$returnValue .= $value["ext"].",";
	}
	return rtrim($returnValue, ",");
}
function getDirName($name){
	$explode = explode("/", $name);
	return end($explode);
}
function getSize($bit,$array=false){
	$type = array('Bytes','KB','MB','GB','TB');  
	$box = array('1','1024','1048576','1073741824','TB');  
	for($i = 0; $bit >= 1024; $i++) {  
		$bit/=1024;  
	}
	if($array){
		return [(floor($bit*100)/100),$box[$i]];  
	}
	return (floor($bit*100)/100).$type[$i];  
}