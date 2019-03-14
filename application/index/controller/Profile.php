<?php
namespace app\index\controller;

use think\Controller;
use app\index\model\User;
use think\Cookie;
use think\Db;
use \app\index\model\Option;

class Profile extends Controller{

	public $visitorObj;
	public $userObj;
	public $siteOptions;

	public function _initialize(){
	}

	public function getList(){
		$userId = (string)input("post.uid");
		$userData = Db::name("users")->where("id",$userId)->find();
		$page = (int)input("post.page");
		if (empty($userId) || empty($userData) || $userData["profile"] == 0){
			$this->error('用户主页不存或者用户关闭了个人主页',404,$this->siteOptions);
		   }
		   switch (input("post.type")) {
			case 'all':
				$list = Db::name('shares')
				->where('owner',$userId)
				->where('type',"public")
				->order('share_time DESC')
				->page($page.',10')
				->select();
				break;
			case 'hot':
				$num = Option::getValue("hot_share_num");
				$list = Db::name('shares')
				->where('owner',$userId)
				->where('type',"public")
				->order('download_num DESC')
				->limit($num)
				->select();
				break;
			default:
				$list = Db::name('shares')
				->where('owner',$userId)
				->where('type',"public")
				->order('share_time DESC')
				->page($page.',10')
				->select();
				break;
		}
		$listData = $list;
		foreach ($listData as $key => $value) {
			unset($listData[$key]["share_pwd"]);
			unset($listData[$key]["source_name"]);
			if($value["source_type"]=="file"){
				$listData[$key]["fileData"] = Db::name('files')->where('id',$value["source_name"])->find()["orign_name"];

			}else{
				$pathDir = explode("/",$value["source_name"]);
				$listData[$key]["fileData"] = end($pathDir);
			}
		}

		return json($listData);


	}

	public function index(){
		$this->visitorObj = new User(cookie('user_id'),cookie('login_key'));
		$this->siteOptions = Option::getValues(["basic"],$this->visitorObj->userSQLData);
		$userId = (string)input("param.uid");
		$userData = Db::name("users")->where("id",$userId)->find();
		if (empty($userId) || empty($userData) || $userData["profile"] == 0){
			 $this->error('用户主页不存或者用户关闭了个人主页',404,$this->siteOptions);
		}
		$groupData = Db::name("groups")->where("id",$userData["user_group"])->find();
		$shareCount = Db::name('shares')
				->where('owner',$userId)
				->where('type',"public")
				->count();
		$regDays = (int)((time()-strtotime($userData["user_date"]))/86400);
		
		return view("profile",[
			"options" => $this->siteOptions,
			'loginStatus' => $this->visitorObj->loginStatus,
			'targetUserInfo' => $userData,
			'userSQL' => $this->visitorObj->userSQLData,
			'userInfo' => $this->visitorObj->getInfo(),
			'groupData' => $groupData,
			'type' => input("get.type"),
			'shareCount' => $shareCount,
			'regDays' => $regDays,
		]);
	}

}
