<?php
namespace app\index\controller;

use think\Controller;
use think\Db;
use \app\index\model\Option;
use \app\index\model\User;
use think\Session;

class Home extends Controller{

	public $userObj;

	public function _initialize(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		if(!$this->userObj->loginStatus){
			$this->redirect(url('/Login','',''));
			exit();
		}
	}

	public function index(){
		$userInfo = $this->userObj->getInfo();
		$policyData = $this->userObj->getPolicy();
		$groupData =  $this->userObj->getGroupData();
		$extJson = json_decode($policyData["filetype"],true);
		$extLimit="";
		foreach ($extJson as $key => $value) {
			$extLimit.='{ title : "'.$value["title"].'", extensions : "'.$value["ext"].'" },';
		}
		$policyData["max_size"] = $policyData["max_size"]/(1024*1024);
		return view('home', [
			'options'  => Option::getValues(['basic','upload']),
			'userInfo' => $userInfo,
			'extLimit' => $extLimit,
			'policyData' => $policyData,
			'groupData' => $groupData,
		]);
	}

	public function Download(){
		$userInfo = $this->userObj->getInfo();
		$groupData =  $this->userObj->getGroupData();
		return view('download', [
			'options'  => Option::getValues(['basic','group_sell']),
			'userInfo' => $userInfo,
			'groupData' => $groupData,
		]);
	}

	public function Album(){
		$userInfo = $this->userObj->getInfo();
		$list = Db::name("files")->where("upload_user",$this->userObj->uid)
					->where(function ($query) {
					    $query->where('orign_name', "like","%jpg")
					    ->whereor('orign_name', "like","%png")
					    ->whereor('orign_name', "like","%gif")
					    ->whereor('orign_name', "like","%bmp");
					})
					->order('id DESC')
					->paginate(9);
		$pageCount = ceil(Db::name("files")->where("upload_user",$this->userObj->uid)
					->where(function ($query) {
					    $query->where('orign_name', "like","%jpg")
					    ->whereor('orign_name', "like","%png")
					    ->whereor('orign_name', "like","%gif")
					    ->whereor('orign_name', "like","%bmp");
					})
					->order('id DESC')->count()/9);
		$listData = $list->all();
		$pageNow = input("?get.page")?input("get.page"):1;
		if($pageNow>$pageCount){
			$this->error('您当前没有上传任何图片',404,Option::getValues(['basic','group_sell']));
		}
		return view('album', [
			'options'  => Option::getValues(['basic','group_sell']),
			'userInfo' => $userInfo,
			'list' => $listData,
			'listOrigin' => $list,
			'pageCount' => $pageCount,
			'page' => $pageNow,
		]);
	}
		
}
