<?php
namespace app\index\controller;

use think\Controller;
use app\index\model\User;
use think\Cookie;
use think\Db;
use \app\index\model\Option;
use \app\index\model\AdminHandler;
use \app\index\model\FileManage;

class Admin extends Controller{

	public $userObj;
	public $siteOptions;
	public $adminObj;

	public function _initialize(){
		$this->siteOptions = Option::getValues(["basic","admin"]);
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		if(!$this->userObj->loginStatus){
			$this->redirect(url('/Login','',''));
			exit();
		}
		if($this->userObj->groupData["id"] != 1){
			$this->error('你无权访问此页面',403,$this->siteOptions);
		}
		$this->adminObj = new AdminHandler($this->siteOptions);
	}

	public function index(){
		if($this->adminObj->checkDbVersion()){
			$this->redirect(url('/Admin/UpdateDb','',''));
			exit();
		}
		return view('admin_index', [
			'options'  => $this->siteOptions,
			'statics' => $this->adminObj->getStatics(),
		]);
	}

	public function UpdateDb(){
		if(!$this->adminObj->checkDbVersion()){
			$this->redirect(url('/Admin','',''));
			exit();
		}
		echo "<meta charset='utf-8'>";
		echo "准备升级数据库，当前数据库版本:".$this->adminObj->dbVerInfo["now"];
		echo ", 目标数据库版本:".$this->adminObj->dbVerInfo["require"].";<br>";
		$updatePath = ROOT_PATH . "update_".$this->adminObj->dbVerInfo["now"]."to".$this->adminObj->dbVerInfo["require"].".sql";
		if(!file_exists($updatePath)){
			die("数据库更新文件(".$updatePath.")不存在，升级中止.");
		}
		echo "获取升级SQL文件(".$updatePath.")<br>";
		$updateContent = file_get_contents($updatePath);
		echo "将执行以下指令：<br>";
		echo "<code>".htmlspecialchars($updateContent)."</code><br>";
		$sqlSingle = explode(";", $updateContent);
		foreach ($sqlSingle as $key => $value){
			if(empty($value)){
				continue;
			}
			if(!Db::execute($value)){
				echo "<strong>执行$value 时出现错误</strong><br>";
			}
		}
		echo "<strong>升级完成,<a href='/Admin'>返回管理面板</a></strong>";
	}

	public function Setting(){
		return view('basic_setting', [
			'options'  => $this->siteOptions,
		]);
	}

	public function Config(){
		$configType=input("?param.type") ? input("param.type") : "common";
		$configFile = $this->adminObj->getConfigFile($configType);
		return view('config_file', [
			'options'  => $this->siteOptions,
			'type' => $configType,
			'content' =>  $configFile[0],
			'path' => $configFile[1],
		]);
	}

	public function SaveConfigFile(){
		return $this->adminObj->saveConfigFile(input('post.'));
	}

	public function SettingReg(){
		return view('reg_setting', [
			'options'  => $this->siteOptions,
			'optionsForSet'  =>  Option::getValues(["login","register"]),
			'groups' => $this->adminObj->getAvaliableGroup(),
		]);
	}

	public function Theme(){
		$fileName=input("?param.name") ? input("param.name") : "error";
		$dir = ROOT_PATH."application/index/view/";
		if(!function_exists("scandir")){
			return "scandir被禁用";
		}
		$fileList=[];
		$fileList=$fileList+scandir($dir);
		$pathList=["/"=>$fileList];
		foreach (["admin","explore","file","home","index","member","profile","share"] as $key => $value) {
			$childPath = scandir($dir.$value."/");
			$fileList=array_merge($fileList,$childPath);
			$pathList = array_merge($pathList,[$value => $childPath]);
		}
		foreach ($fileList as $key => $value) {
			if(substr_compare($value, ".html", -strlen(".html")) != 0){
				unset($fileList[$key]);
			}
		}
		foreach($pathList as $key=>$val){
		    if(in_array($fileName.".html",$val)){
		        $parentPath = $key;
		        break;
		    }
		}
		$fileContent = file_get_contents($dir.rtrim($parentPath,"/")."/".$fileName.".html");
		return view('theme', [
			'options'  => $this->siteOptions,
			'list' => $fileList,
			'content' =>  $fileContent,
			'path' => $parentPath,
			'name' => $fileName,
		]);
	}

	public function SaveThemeFile(){
		return $this->adminObj->saveThemeFile(input('post.'));
	}

	public function SettingMail(){
		return view('mail_setting', [
			'options'  => $this->siteOptions,
			'optionsForSet'  =>  Option::getValues(["mail_template","mail"]),
		]);
	}

	public function SettingPay(){
		return view('pay_setting', [
			'options'  => $this->siteOptions,
			'optionsForSet'  =>  Option::getValues(["payment"]),
		]);
	}

	public function SettingOther(){
		return view('other_setting', [
			'options'  => $this->siteOptions,
			'optionsForSet'  =>  Option::getValues(["file_edit","share","avatar","admin","storage_policy","download"]),
		]);
	}

	public function Files(){
		$this->adminObj->listFile();
		return view('file_list', [
			'options'  => $this->siteOptions,
			'groups' => $this->adminObj->getAvaliableGroup(),
			'list' => $this->adminObj->pageData,
			'originList' => $this->adminObj->listData,
			'pageNow' => $this->adminObj->pageNow,
			'pageTotal' => $this->adminObj->pageTotal,
			'dataTotal' => $this->adminObj->dataTotal,
			'policy' => $this->adminObj->getAvaliablePolicy(),
		]);
	}

	Public function Users(){
		$this->adminObj->listUser();
		$group = $this->adminObj->getAvaliableGroup();
		return view('user_list', [
			'options'  => $this->siteOptions,
			'group' => $group,
			'groups' => $group,
			'list' => $this->adminObj->pageData,
			'originList' => $this->adminObj->listData,
			'pageNow' => $this->adminObj->pageNow,
			'pageTotal' => $this->adminObj->pageTotal,
			'dataTotal' => $this->adminObj->dataTotal,
			'policy' => $this->adminObj->getAvaliablePolicy(),
		]);
	}

	public function Shares(){
		$this->adminObj->listShare();
		return view('share_list', [
			'options'  => $this->siteOptions,
			'groups' => $this->adminObj->getAvaliableGroup(),
			'list' => $this->adminObj->pageData,
			'originList' => $this->adminObj->listData,
			'pageNow' => $this->adminObj->pageNow,
			'pageTotal' => $this->adminObj->pageTotal,
			'dataTotal' => $this->adminObj->dataTotal,
		]);
	}

	public function PolicyList(){
		$this->adminObj->listPolicy();
		return view('policy_list', [
			'options'  => $this->siteOptions,
			'groups' => $this->adminObj->getAvaliableGroup(),
			'list' => $this->adminObj->pageData,
			'originList' => $this->adminObj->listData,
			'pageNow' => $this->adminObj->pageNow,
			'pageTotal' => $this->adminObj->pageTotal,
			'dataTotal' => $this->adminObj->dataTotal,
		]);
	}

	public function GroupList(){
		$this->adminObj->listGroup();
		return view('group_list', [
			'options'  => $this->siteOptions,
			'list' => $this->adminObj->pageData,
			'originList' => $this->adminObj->listData,
			'pageNow' => $this->adminObj->pageNow,
			'pageTotal' => $this->adminObj->pageTotal,
			'dataTotal' => $this->adminObj->dataTotal,
		]);
	}

	public function OrderList(){
		$this->adminObj->listOrder();
		return view('order_list', [
			'options'  => $this->siteOptions,
			'list' => $this->adminObj->pageData,
			'originList' => $this->adminObj->listData,
			'pageNow' => $this->adminObj->pageNow,
			'pageTotal' => $this->adminObj->pageTotal,
			'dataTotal' => $this->adminObj->dataTotal,
		]);
	}

	public function SaveBasicSetting(){
		return $this->adminObj->saveBasicSetting(input('post.'));
	}

	public function SaveRegSetting(){
		return $this->adminObj->saveRegSetting(input('post.'));
	}

	public function SaveMailSetting(){
		return $this->adminObj->saveMailSetting(input('post.'));
	}

	public function SaveAria2Setting(){
		return $this->adminObj->saveAria2Setting(input('post.'));
	}

	public function SendTestMail(){
		return $this->adminObj->sendTestMail(input('post.'));
	}

	public function SaveMailTemplate(){
		return $this->adminObj->saveMailTemplate(input('post.'));
	}

	public function GetFileInfo(){
		return $this->adminObj->getFileInfo(input('post.id'));
	}

	public function GetUserInfo(){
		return $this->adminObj->getUserInfo(input('post.id'));
	}

	public function savePolicy(){
		return $this->adminObj->addPolicy(input('post.'));
	}

	public function SaveEditPolicy(){
		return $this->adminObj->editPolicy(input('post.'));
	}

	public function SaveGroup(){
		return $this->adminObj->saveGroup(input('post.'));
	}

	public function AddPack(){
		return $this->adminObj->addPack(input('post.'));
	}

	public function AddGroupPurchase(){
		return $this->adminObj->addGroupPurchase(input('post.'));
	}

	public function SaveCron(){
		$this->adminObj->saveCron(input('post.'));
		$this->redirect("/Admin/Cron",302);
	}

	public function SaveUser(){
		return $this->adminObj->saveUser(input('post.'));
	}

	public function BanUser(){
		return $this->adminObj->banUser(input('post.id'),$this->userObj->uid);
	}

	public function AddUser(){
		return $this->adminObj->addUser(input('post.'));
	}

	public function Preview(){
		$fileId = input('param.id');
		$fileRecord = Db::name("files")->where("id",$fileId)->find();
		$fileObj = new FileManage(rtrim($fileRecord["dir"],"/")."/".$fileRecord["orign_name"],$fileRecord["upload_user"]);
		$previewHandler = $fileObj->PreviewHandler(true);
		if($previewHandler[0]){
			$this->redirect($previewHandler[1],302);
		}
	}

	public function Download(){
		$fileId = input('param.id');
		$fileRecord = Db::name("files")->where("id",$fileId)->find();
		$fileObj = new FileManage(rtrim($fileRecord["dir"],"/")."/".$fileRecord["orign_name"],$fileRecord["upload_user"]);
		$FileHandler = $fileObj->Download(true);
		if($FileHandler[0]){
			$this->redirect($FileHandler[1],302);
		}
	}

	public function Delete(){
		return $this->adminObj->deleteSingle(input('post.id'));
	}

	public function DeleteShare(){
		return $this->adminObj->deleteShare([0=>input('post.id')]);
	}

	public function DeleteShareMultiple(){
		return $this->adminObj->deleteShare(json_decode(input('post.id'),true));
	}

	public function DeleteMultiple(){
		return $this->adminObj->deleteMultiple(input('post.id'));
	}

	public function DeletePolicy(){
		return $this->adminObj->deletePolicy(input('post.id'));
	}

	public function DeleteGroup(){
		return $this->adminObj->deleteGroup(input('post.id'));
	}

	public function DeleteOrder(){
		return $this->adminObj->deleteOrder(input('post.id'));
	}

	public function ChangeShareType(){
		return $this->adminObj->changeShareType(input('post.id'));
	}

	public function DeletePack(){
		return $this->adminObj->deletePack(input('post.id'));
	}

	public function DeleteGroupPurchase(){
		return $this->adminObj->deleteGroupPurchase(input('post.id'));
	}

	public function DeleteUser(){
		return $this->adminObj->deleteUser(input('post.id'),$this->userObj->uid);
	}

	public function DeleteUsers(){
		$uidGroup = json_decode(input('post.id'),true);
		foreach ($uidGroup as $key => $value) {
			$this->adminObj->deleteUser($value,$this->userObj->uid);
		}
		return ["error"=>false,"msg"=>"删除成功"];
	}

	public function SwitchColor(){
		$colorNow = Option::getValues(["admin"]);
		if($colorNow["admin_color_body"] == "fixed-nav sticky-footer bg-light"){
			$colorNew = [
				"admin_color_body" => "fixed-nav sticky-footer bg-dark",
				"admin_color_nav" => "navbar navbar-expand-lg fixed-top navbar-dark bg-dark",
			];
		}else{
			$colorNew = [
				"admin_color_body" => "fixed-nav sticky-footer bg-light",
				"admin_color_nav" => "navbar navbar-expand-lg fixed-top navbar-light bg-light",
			];
		}
		foreach ($colorNew as $key => $value) {
			Db::name("options")->where("option_name",$key)->update(["option_value" => $value]);
		}
	}

	public function EditPolicy(){
		$policyId = input('param.id');
		$policyRecord = Db::name("policy")->where("id",$policyId)->find();
		return view('edit_policy', [
			'options' => $this->siteOptions,
			'policy' => $policyRecord,
		]);
	}

	public function EditGroup(){
		$groupId = input('param.id');
		$groupRecord = Db::name("groups")->where("id",$groupId)->find();
		return view('edit_group', [
			'options' => $this->siteOptions,
			'group' => $groupRecord,
			'policy' => $this->adminObj->getAvaliablePolicy(),
		]);
	}

	public function AddGroup(){
		return $this->adminObj->addGroup(input('post.'));
	}

	public function PolicyAdd(){
		return view('add_policy', [
			'options' => $this->siteOptions,
		]);
	}

	public function Cron(){
		$cronData = Db::name("corn")->select();
		$neverExcute = true;
		foreach ($cronData as $key => $value) {
			if($value["last_excute"] !=0){
				$neverExcute = false;
			}
		}
		return view('cron_list', [
			'options' => $this->siteOptions,
			'cron' => $cronData,
			'neverExcute' => $neverExcute,
		]);
	}

	public function PolicyAddS3(){
		return view('add_policy_s3', [
			'options' => $this->siteOptions,
		]);
	}

	public function PolicyAddRemote(){
		return view('add_policy_remote', [
			'options' => $this->siteOptions,
		]);
	}

	public function PolicyAddOnedrive(){
		return view('add_policy_onedrive', [
			'options' => $this->siteOptions,
		]);
	}

	public function About(){
		$verison = json_decode(file_get_contents(ROOT_PATH . "application/version.json"),true);
		return view('about', [
			'options' => $this->siteOptions,
			'programVersion' => $verison,
			"dbsVersion" => Option::getValue("database_version"),
		]);
	}

	public function Purchase(){
		$packData = json_decode(Option::getValue("pack_data"),true);
		return view('purchase', [
			'options' => $this->siteOptions,
			'pack' => $packData,
		]);
	}

	public function PurchaseGroup(){
		$groupData = json_decode(Option::getValue("group_sell_data"),true);
		foreach ($groupData as $key => $value) {
			$groupData[$key]["group"] = Db::name("groups")->where("id",$value["goup_id"])->find();
		}
		return view('purchase_group', [
			'options' => $this->siteOptions,
			'group' => $groupData,
			'group_list' => $this->adminObj->getAvaliableGroup(),
		]);
	}

	public function GroupAdd(){
		return view('add_group', [
			'options' => $this->siteOptions,
			'policy' => $this->adminObj->getAvaliablePolicy(),
		]);
	}

	public function RemoteDownload(){
		$this->adminObj->listDownloads();
		return view('download', [
			'options'  => $this->siteOptions,
			'optionsForSet'  =>  Option::getValues(["aria2"]),
			'list' => $this->adminObj->pageData,
			'originList' => $this->adminObj->listData,
			'pageNow' => $this->adminObj->pageNow,
			'pageTotal' => $this->adminObj->pageTotal,
			'dataTotal' => $this->adminObj->dataTotal,
		]);
	}

	public function CancelDownload(){
		$aria2Options = Option::getValues(["aria2"]);
		$aria2 = new \app\index\model\Aria2($aria2Options);
		$downloadItem =  Db::name("download")->where("id",input("post.id"))->find();
		if(empty($downloadItem)){
			return json(['error'=>1,'message'=>"未找到下载记录"]);
		}
		if($aria2->Remove($downloadItem["pid"],"")){
			return json(['error'=>0,'message'=>"下载已取消"]);
		}else{
			return json(['error'=>1,'message'=>"取消失败"]);
		}
	}

	public function UpdateOnedriveToken(){
		$policyId = input("get.id");
		$this->adminObj->updateOnedriveToken($policyId);

	}

	public function OneDriveCalllback(){
		$code = input("get.code");
		$this->adminObj->oneDriveCalllback($code);
	}
	
}
