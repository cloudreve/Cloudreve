<?php
namespace app\index\model;

use think\Model;
use think\Db;
use app\index\model\User;

class DavAuth extends Model{

	public $uid;

	function __construct($id) {
		$this->uid = $id;
	}

	public function  __invoke($realm,$um){
		$userData = Db::name("users")->where("id",$this->uid)->find();
		if(empty($userData) || $userData["user_email"] != $um){
			return null;
		}
		$userGroup = Db::name("groups")->where("id",$userData["user_group"])->find();
		if(!$userGroup["webdav"]){
			return null;
		}
		return $userData["webdav_key"];
	}

}

?>