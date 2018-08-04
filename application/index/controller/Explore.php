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
		$this->siteOptions = Option::getValues(["basic"]);
	}

	public function Search(){
		$this->visitorObj = new User(cookie('user_id'),cookie('login_key'));
		return view("search",[
			"options" => $this->siteOptions,
			'loginStatus' => $this->visitorObj->loginStatus,
			'userData' => $this->visitorObj->userSQLData,
		]);
	}

	public function S(){
		$this->visitorObj = new User(cookie('user_id'),cookie('login_key'));
		$keyWords=input("param.key");
		if(empty($keyWords)){
			$this->redirect('/Explore/Search',302);
		}
		$list = Db::name('shares')
				->where('type',"public")
				->where('origin_name',"like","%".$keyWords."%")
				->order('share_time DESC')
				->paginate(10);
		$listData = $list->all();
		foreach ($listData as $key => $value) {
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
			'userData' => $this->visitorObj->userSQLData,
			'list' => $listData,
			'listOrigin' => $list,
			'keyWords' => $keyWords,
		]);
	}

}
