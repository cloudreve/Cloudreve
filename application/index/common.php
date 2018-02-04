<?php
function countSize($bit,$array=false){
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

function isPreview($fileName){
	$allowedSuffix=["jpg","jpeg","gif","bmp","png","svg","mp4","mp3","ogg"];
	$suffix = explode(".",$fileName);
	if(in_array(end($suffix),$allowedSuffix)){
		return "yes";
	}
	return "no";
}
function getDay($sec){
	return floor($sec/86400);
}
?>