<?php
namespace app\index\controller;

use think\Controller;
use app\index\model\User;
use think\Cookie;
use think\Db;
use \app\index\model\Option;

class Explore extends Controller{

	public $visitorObj;
	public $userObj;
	public $siteOptions;

	public function _initialize(){
	}

	public function Search(){
		$this->visitorObj = new User(cookie('user_id'),cookie('login_key'));
		$this->siteOptions = Option::getValues(["basic"],$this->visitorObj->userSQLData);
		$keyWords=input("param.key");
		if(empty($keyWords)){
			$this->error("搜索词不为空",200,$this->siteOptions);
		}else{
			$list = Db::name('shares')
				->where('type',"public")
				->where('origin_name',"like","%".$keyWords."%")
				->order('share_time DESC')
				->select();
		}
		$listData = $list;
		foreach ($listData as $key => $value) {
			unset($listData[$key]["source_name"]);
			if($value["source_type"]=="file"){
				$listData[$key]["fileData"] = $value["origin_name"];

			}else{
				$pathDir = explode("/",$value["source_name"]);
				$listData[$key]["fileData"] = end($pathDir);
			}
		}
		return view("result",[
			"options" => $this->siteOptions,
			'loginStatus' => $this->visitorObj->loginStatus,
			'userData' => $this->visitorObj->getInfo(),
			'list' => json_encode($listData),
			'keyWords' => $keyWords,
		]);
	}

}
