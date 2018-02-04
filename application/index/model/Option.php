<?php
namespace app\index\model;

use think\Model;
use think\Db;

class Option extends Model{
	static function getValues($groups = ['basic']){
		$t =  Db::name('options')->where('option_type','in',$groups)->column('option_value','option_name');
		return $t;
	}
	static function getValue($optionName){
		return Db::name('options')->where('option_name',$optionName)->value('option_value');
	}
}
?>