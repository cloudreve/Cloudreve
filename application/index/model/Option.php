<?php
namespace app\index\model;

use think\Model;
use think\Db;

class Option extends Model{
	
	static function getValues($groups = ['basic'],$userInfo=null){
		$t =  Db::name('options')->where('option_type','in',$groups)->column('option_value','option_name');
		if(in_array("basic",$groups)){
			return array_merge($t,self::getThemeOptions($t,$userInfo));
		}
		return $t;
	}

	static function getThemeOptions($basicOptions,$userInfo){
		$themes = json_decode($basicOptions["themes"],true);
		if($userInfo==null){
			return ["themeColor"=>$basicOptions["defaultTheme"],"themeConfig"=>$themes[$basicOptions["defaultTheme"]]];
		}else{
			$userOptions = json_decode($userInfo["options"],true);
			if(empty($userOptions)||!array_key_exists("preferTheme",$userOptions)||!array_key_exists($userOptions["preferTheme"],$themes)){
				return ["themeColor"=>$basicOptions["defaultTheme"],"themeConfig"=>$themes[$basicOptions["defaultTheme"]]];
			}
			return ["themeColor"=>$userOptions["preferTheme"],"themeConfig"=>$themes[$userOptions["preferTheme"]]];
		}
	}

	static function getValue($optionName){
		return Db::name('options')->where('option_name',$optionName)->value('option_value');
	}
}
?>