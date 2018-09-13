<?php
namespace app\index\model;

use think\Model;
use think\Db;
use \app\index\model\Mail;
use \app\index\model\FileManage;
use \Krizalys\Onedrive\Client;

class AdminHandler extends Model{

	public $siteOptions;
	public $pageData;
	public $listData;
	public $pageNow;
	public $pageTotal;
	public $dataTotal;
	public $dbVerInfo;
	
	public function __construct($options){
		$this->siteOptions = $options;
	}

	public function checkDbVersion(){
		$versionInfo = json_decode(@file_get_contents(ROOT_PATH. "application/version.json"),true);
		$dbVerNow = Option::getValue("database_version");
		if(!isset($versionInfo["db_version"]) || $dbVerNow < (int)$versionInfo["db_version"]){
			$this->dbVerInfo = array('now' => $dbVerNow, 'require' => $versionInfo["db_version"]);
			return true;
		}
		return false;
	}

	public function getStatics(){
		$statics["fileNum"] = Db::name('files')->count();
		$statics["privateShareNum"] = Db::name('shares')->where("type","private")->count();
		$statics["publicShareNum"] = Db::name('shares')->where("type","public")->count();
		$statics["userNum"] = Db::name('users')->where("user_status",0)->count();
		if($statics["fileNum"]==0){
			$statics["imgRate"] = 0;
			$statics["audioRate"] = 0;
			$statics["videoRate"] = 0;
			$statics['otherRate'] = 0;
		}else{
			$statics["imgRate"] =floor(Db::name('files')
			->where('pic_info',"<>"," ")
			->where('pic_info',"<>","0,0")
			->where('pic_info',"<>","null,null")
			->count()/$statics["fileNum"]*10000)/100;
			$statics["audioRate"] =floor(Db::name('files')
			->where(function ($query) {
			    $query->where('orign_name', "like","%mp3")
			    ->whereor('orign_name', "like","%flac")
			    ->whereor('orign_name', "like","%wma")
			    ->whereor('orign_name', "like","%aac")
			    ->whereor('orign_name', "like","%wav")
			    ->whereor('orign_name', "like","%ogg");
			})
			->count()/$statics["fileNum"]*10000)/100;
			$statics["videoRate"] =floor(Db::name('files')
			->where(function ($query) {
			    $query->where('orign_name', "like","%mp4")
			    ->whereor('orign_name', "like","%avi")
			    ->whereor('orign_name', "like","%rmvb")
			    ->whereor('orign_name', "like","%aac")
			    ->whereor('orign_name', "like","%wav")
			    ->whereor('orign_name', "like","%mkv");
			})
			->count()/$statics["fileNum"]*10000)/100;
			$statics['otherRate'] = 100-($statics["videoRate"]+$statics["audioRate"]+$statics["imgRate"]);
		}
		$timeNow=time();
		$statics["trendFile"]="";
		$statics["trendUser"]="";
		$statics["trendDate"]="";
		for ($i=0; $i < 13; $i++) { 
			$statics["trendFile"].= Db::name('files')->where('upload_date','between time',[date("Y-m-d",$timeNow-(12-$i)*3600*24),date("Y-m-d",$timeNow-(11-$i)*3600*24)])->count().",";
		}
		for ($i=0; $i < 13; $i++) { 
			$statics["trendUser"].= Db::name('users')->where('user_date','between time',[date("Y-m-d",$timeNow-(12-$i)*3600*24),date("Y-m-d",$timeNow-(11-$i)*3600*24)])->count().",";
		}
		for ($i=0; $i < 13; $i++) { 
			$statics["trendDate"].='"'.date("m月d日",$timeNow-(12-$i)*3600*24).'",';
		}
		$statics["trendFile"] = rtrim($statics["trendFile"],",");
		$statics["trendDate"] = rtrim($statics["trendDate"],",");
		$statics["trendUser"] = rtrim($statics["trendUser"],",");
		return $statics;
	}

	public function saveBasicSetting($options){
		$siteUrl = rtrim($options["siteURL"],"/")."/";
		$options["siteURL"]=$siteUrl;
		return $this->saveOptions($options);
	}

	public function saveRegSetting($options){
		foreach(["email_active","login_captcha","reg_captcha","forget_captcha"] as $key){
			$options[$key] = array_key_exists($key,$options) ? $options[$key] : 0;
		}
		return $this->saveOptions($options);
	}

	public function saveMailSetting($options){
		return $this->saveOptions($options);
	}

	public function saveAria2Setting($options){
		return $this->saveOptions($options);
	}

	public function saveMailTemplate($options){
		return $this->saveOptions($options);
	}

	public function AddGroup($options){
		$options["max_storage"] = $options["max_storage"]*$options["sizeTimes"];
		unset($options["sizeTimes"]);
		$options["grade_policy"] = 0;
		$options["policy_list"] = $options["policy_name"];
		$options["aria2"] = $options["aria2"] ? "1,1,1" : "0,0,0";
		try {
			Db::name("groups")->insert($options);
		} catch (Exception $e) {
			return ["error"=>1,"msg"=>$e->getMessage()];
		}
		return ["error"=>200,"msg"=>"设置已保存"];
	}

	public function addPolicy($options){
		$options["max_size"] = $options["max_size"]*$options["sizeTimes"];
		unset($options["sizeTimes"]);
		$options["server"] = isset($options["server"]) ? $options["server"] : "/Upload";
		foreach (["bucketname","bucket_private","bucketname","ak","sk","op_name","op_pwd","mimetype","namerule"] as $key => $value) {
			$options[$value] = isset($options[$value]) ? $options[$value] : "0";
		}
		if(empty($options["filetype"])){
			$options["filetype"]="[]";
		}else{
			$options["filetype"] = json_encode([0=>["ext"=>$options["filetype"],"title"=>"default"]]);
		}
		if($options["policy_type"] == "upyun"){
			$options["server"] = "https://v0.api.upyun.com/".$options["bucketname"];
		}
		try {
			Db::name("policy")->insert($options);
		} catch (Exception $e) {
			return ["error"=>1,"msg"=>$e->getMessage()];
		}
		return ["error"=>200,"msg"=>"设置已保存","id"=>Db::name('policy')->getLastInsID()];
	}

	public function editPolicy($options){
		$policyId = $options["id"];
		$options["max_size"] = $options["max_size"]*$options["sizeTimes"];
		unset($options["sizeTimes"]);
		unset($options["id"]);
		if(empty($options["filetype"])){
			$options["filetype"]="[]";
		}else{
			$options["filetype"] = json_encode([0=>["ext"=>$options["filetype"],"title"=>"default"]]);
		}
		try {
			Db::name("policy")->where("id",$policyId)->update($options);
		} catch (Exception $e) {
			return ["error"=>1,"msg"=>$e->getMessage()];
		}
		return ["error"=>200,"msg"=>"设置已保存"];
	}

	public function saveGroup($options){
		$groupId = $options["id"];
		unset($options["id"]);
		$options["max_storage"] = $options["max_storage"]*$options["sizeTimes"];
		unset($options["sizeTimes"]);
		$options["aria2"] = $options["aria2"] ? "1,1,1" : "0,0,0";
		try {
			Db::name("groups")->where("id",$groupId)->update($options);
		} catch (Exception $e) {
			return ["error"=>1,"msg"=>$e->getMessage()];
		}
		return ["error"=>200,"msg"=>"设置已保存"];
	}

	public function saveOptions($options){
		try {
			foreach ($options as $key => $value) {
				Db::name("options")->where("option_name",$key)->update(["option_value"=>$value]);
			}
		} catch (Exception $e) {
			return ["error"=>1,"msg"=>$e->getMessage()];
		}
		return ["error"=>200,"msg"=>"设置已保存"];
	}

	public function getAvaliableGroup(){
		$groupData = Db::name("groups")->where("id","neq",2)->select();
		return $groupData;
	}

	public function saveCron($options){
		$cronId = $options["id"];
		unset($options["id"]);
		Db::name("corn")->where("id",$cronId)->update($options);
	}

	public function getAvaliablePolicy(){
		$policyData = Db::name("policy")->select();
		return $policyData;
	}

	public function sendTestMail($options){
		$mailObj = new Mail();
		if(empty($options["receiveMail"])){
			return ["error"=>1,"msg"=>"接收邮箱不能为空"];
		}
		$sendResult = $mailObj->Send($options["receiveMail"],"发信测试",$options["subject"],$options["content"]);
		if($sendResult){
			return ["error"=>200,"msg"=>"发送成功"];
		}else{
			return ["error"=>1,"msg"=>$mailObj->errorMsg];
		}
	}

	public function deleteSingle($id){
		$fileRecord = Db::name("files")->where("id",$id)->find();
		return FileManage::DeleteHandler([0 => rtrim($fileRecord["dir"],"/")."/".$fileRecord["orign_name"]],$fileRecord["upload_user"]);
	}

	public function deletePolicy($id){
		$groupData = Db::name("groups")->where("policy_name",$id)->select();
		if(!empty($groupData)){
			return ["error"=>true,"msg"=>"此上传策略正在被以下用户组使用：".join(",",array_column($groupData, "group_name"))];
		}
		Db::name("policy")->where("id",$id)->delete();
		return ["error"=>false,"msg"=>"已删除"];
	}

	public function deleteGroup($id){
		$userData = Db::name("users")->where("user_group",$id)->find();
		if(!empty($userData)){
			return ["error"=>true,"msg"=>"此用户组下仍有用户，请先删除这些用户"];
		}
		if($id == 1 || $id == 2){
			return ["error"=>true,"msg"=>"系统保留用户组，无法删除"];
		}
		Db::name("groups")->where("id",$id)->delete();
		return ["error"=>false,"msg"=>"已删除"];
	}

	public function getConfigFile($type){
		switch ($type) {
			case 'common':
				$configPath = ROOT_PATH ."application/config.php"; 
				$basicPath = "application/config.php";
				break;
			case 'database':
				if(file_exists( ROOT_PATH ."application/database.lock")){
					return ["出于安全考虑，默认禁止直接编辑数据库配置文件。如果需要开启编辑，请手动删除 application/database.lock 文件。","application/database.php"];
				}
				$configPath = ROOT_PATH ."application/database.php"; 
				$basicPath = "application/database.php";
				break;
			case 'route':
				$configPath = ROOT_PATH ."application/route.php"; 
				$basicPath = "application/route.php";
				break;
			case 'tags':
				$configPath = ROOT_PATH ."application/tags.php"; 
				$basicPath = "application/tags.php";
				break;
			default:
				die("");
				break;
		}
		return [file_get_contents($configPath),$basicPath];
	}

	public function saveConfigFile($options){
		switch ($options["type"]) {
			case 'common':
				file_put_contents(ROOT_PATH ."application/config.php",$options["content"]);
				break;
			case 'route':
				file_put_contents(ROOT_PATH ."application/route.php",$options["content"]);
				break;
			case 'tags':
				file_put_contents(ROOT_PATH ."application/tags.php",$options["content"]);
				break;
			case 'database':
				if(file_exists( ROOT_PATH ."application/database.lock")){
					return ["error"=>true,"msg"=>"出于安全考虑，默认禁止直接编辑数据库配置文件。如果需要开启编辑，请手动删除 application/database.lock 文件。"];
				}
				file_put_contents(ROOT_PATH ."application/database.php",$options["content"]);
				break;
			default:
				# code...
				break;
		}
		return ["error"=>false,"msg"=>""];
	}

	public function deleteMultiple($id){
		$fileInfo = json_decode($id,true);
		$pathGroup = [];
		foreach ($fileInfo as $key => $value) {
			$pathGroup[$value["uid"]] = isset($pathGroup[$value["uid"]]) ? $pathGroup[$value["uid"]] : [];
			array_push($pathGroup[$value["uid"]], $value["path"]);
		}
		foreach ($pathGroup as $key => $value) {
			FileManage::DeleteHandler($value,$key);
		}
		return ["error"=>200,"msg"=>"删除成功"];
	}

	public function deleteShare($ids){
		Db::name("shares")->where("id","in",$ids)->delete();
		return ["error"=>false,"msg"=>"删除成功"];
	}

	public function deleteOrder($id){
		Db::name("order")->where("id",$id)->delete();
		return ["error"=>false,"msg"=>"删除成功"];
	}

	public function deleteUser($id,$userNow){
		if($userNow == $id){
			return ["error"=>true,"msg"=>"我的老伙计，你可不能删除你自己"];
		}
		//删除用户所有文件及目录
		FileManage::DirDeleteHandler([0 => "/"],$id);
		//删除此用户所有分享
		Db::name("shares")->where("owner",$id)->delete();
		//删除此用户
		Db::name("users")->where("id",$id)->delete();
		return ["error"=>false,"msg"=>"删除成功"];
	}

	public function changeShareType($id){
		$shareId = $id;
		$shareObj = new ShareHandler($shareId,false);
		if(!$shareObj->querryStatus){
			 return array(
				"error" => 1,
				"msg" => "分享不存在"
				);
		}
		return $shareObj->changePromission(0,true);
	}

	public function getUserInfo($id){
		$userData = Db::name("users")->where("id",$id)->find();
		$userData["used_storage"] =getSize($userData["used_storage"]);
		return $userData;
	}

	public function listDownloads(){
		$pageSize = 10;
		$this->pageData = Db::name("download")
		->order("id desc")
		->paginate($pageSize);
		$this->dataTotal = Db::name("download")
		->order("id desc")
		->count();
		$this->pageTotal = ceil($this->dataTotal/$pageSize);
		$this->listData = $this->pageData->all();
		$userCache=[];
		$userCacheList=[];
		foreach ($this->listData as $key => $value) {
			if(in_array($value["owner"], $userCacheList)){
				$this->listData[$key]["user"] = $userCache[$value["owner"]];
			}else{
				$this->listData[$key]["user"] = Db::name("users")->where("id",$value["owner"])->find();
				array_push($userCacheList,$value["owner"]);
				$userCache[$value["owner"]] = $this->listData[$key]["user"];
			}
			$connectInfo = json_decode($value["info"],true);
			if(isset($connectInfo["dir"])){
				$this->listData[$key]["fileName"] = basename($connectInfo["dir"]);
				$this->listData[$key]["completedLength"] = $connectInfo["completedLength"];
				$this->listData[$key]["totalLength"] = $connectInfo["totalLength"];
				$this->listData[$key]["downloadSpeed"] = $connectInfo["downloadSpeed"];
			}else{
				if(floor($value["source"])==$value["source"]){
					$this->listData[$key]["fileName"] = Db::name("files")->where("id",$value["source"])->column("orign_name");
				}else{
					$this->listData[$key]["fileName"] = $value["source"];
				}
				$this->listData[$key]["completedLength"] = 0;
				$this->listData[$key]["totalLength"] = 0;
				$this->listData[$key]["downloadSpeed"] = 0;
			}
		}
		$this->pageNow = input("?get.page")?input("get.page"):1;
	}

	public function listFile(){
		$pageSize = !cookie('?pageSize') ? 10 : cookie('pageSize');
		$orderType = empty(cookie('orderMethodFile')) ? "id DESC" : cookie('orderMethodFile');
		$this->pageData = Db::name("files")
		->where(function ($query) {
			if(!empty(cookie('fileSearch'))){
			    $query->where('orign_name', "like","%".cookie('fileSearch')."%");
			}
		})
		->where(function ($query) {
			if(!empty(cookie('filePolicy'))){
			    $query->where('policy_id', cookie('filePolicy'));
			}
		})
		->where(function ($query) {
			if(!empty(cookie('searchValue'))){
			    $query->where(cookie('searchCol'),"like", cookie('searchValue'));
			}
		})
		->order($orderType)
		->paginate($pageSize);
		$this->dataTotal = Db::name("files")
		->where(function ($query) {
			if(!empty(cookie('fileSearch'))){
			    $query->where('orign_name', "like","%".cookie('fileSearch')."%");
			}
		})
		->where(function ($query) {
			if(!empty(cookie('filePolicy'))){
			    $query->where('policy_id', cookie('filePolicy'));
			}
		})
		->where(function ($query) {
			if(!empty(cookie('searchValue'))){
			    $query->where(cookie('searchCol'),"like", cookie('searchValue'));
			}
		})
		->order($orderType)
		->count();
		$this->pageTotal = ceil($this->dataTotal/$pageSize);
		$this->listData = $this->pageData->all();
		$userCache=[];
		$userCacheList=[];
		foreach ($this->listData as $key => $value) {
			if(in_array($value["upload_user"], $userCacheList)){
				$this->listData[$key]["user"] = $userCache[$value["upload_user"]];
			}else{
				$this->listData[$key]["user"] = Db::name("users")->where("id",$value["upload_user"])->find();
				array_push($userCacheList,$value["upload_user"]);
				$userCache[$value["upload_user"]] = $this->listData[$key]["user"];
			}
		}
		$this->pageNow = input("?get.page")?input("get.page"):1;
	}

	public function listUser(){
		$pageSize = !cookie('?pageSize') ? 10 : cookie('pageSize');
		$orderType = empty(cookie('orderMethodUser')) ? "id DESC" : cookie('orderMethodUser');
		$this->pageData = Db::name("users")
		->where(function ($query) {
			if(!empty(cookie('userStatus'))){
			    $query->where('user_status', cookie('userStatus')-1);
			}
		})
		->where(function ($query) {
			if(!empty(cookie('userSearch'))){
			    $query->where('user_nick', "like","%".cookie('userSearch')."%")
			    ->whereOr("id",cookie('userSearch'))
			    ->whereOr("user_email","like","%".cookie('userSearch')."%");
			}
		})
		->where(function ($query) {
			if(!empty(cookie('userGroup'))){
			    $query->where('user_group', cookie('userGroup'));
			}
		})
		->where(function ($query) {
			if(!empty(cookie('searchValueUser'))){
			    $query->where(cookie('searchColUser'),"like", cookie('searchValueUser'));
			}
		})
		->order($orderType)
		->paginate($pageSize);
		$this->dataTotal = Db::name("users")
		->where(function ($query) {
			if(!empty(cookie('userStatus'))){
			    $query->where('user_status', cookie('userStatus')-1);
			}
		})
		->where(function ($query) {
			if(!empty(cookie('userSearch'))){
			    $query->where('user_nick', "like","%".cookie('userSearch')."%")
			    ->whereOr("id",cookie('userSearch'))
			    ->whereOr("user_email","like","%".cookie('userSearch')."%");
			}
		})
		->where(function ($query) {
			if(!empty(cookie('userGroup'))){
			    $query->where('user_group', cookie('userGroup'));
			}
		})
		->where(function ($query) {
			if(!empty(cookie('searchValueUser'))){
			    $query->where(cookie('searchColUser'),"like", cookie('searchValueUser'));
			}
		})
		->order($orderType)
		->count();
		$this->pageTotal = ceil($this->dataTotal/$pageSize);
		$this->listData = $this->pageData->all();
		$groupCache=[];
		$groupCacheList=[];
		foreach ($this->listData as $key => $value) {
			if(in_array($value["user_group"], $groupCacheList)){
				$this->listData[$key]["group"] = $groupCache[$value["user_group"]];
			}else{
				$this->listData[$key]["group"] = Db::name("groups")->where("id",$value["user_group"])->find();
				array_push($groupCacheList,$value["user_group"]);
				$groupCache[$value["user_group"]] = $this->listData[$key]["group"];
			}
		}
		$this->pageNow = input("?get.page")?input("get.page"):1;
	}

	public function listShare(){
		$pageSize = !cookie('?pageSize') ? 10 : cookie('pageSize');
		$orderType = empty(cookie('orderMethodShare')) ? "share_time DESC" : cookie('orderMethodShare');
		$this->pageData = Db::name("shares")
		->where(function ($query) {
			if(!empty(cookie('shareSearch'))){
			    $query->where('source_name', "like","%".cookie('shareSearch')."%")->whereOr('origin_name', "like","%".cookie('shareSearch')."%");
			}
		})
		->where(function ($query) {
			if(!empty(cookie('shareType'))){
			    $query->where('type', cookie('shareType'));
			}
		})
		->order($orderType)
		->paginate($pageSize);
		$this->dataTotal = Db::name("shares")
		->where(function ($query) {
			if(!empty(cookie('shareSearch'))){
			    $query->where('source_name', "like","%".cookie('shareSearch')."%")->whereOr('origin_name', "like","%".cookie('shareSearch')."%");
			}
		})
		->where(function ($query) {
			if(!empty(cookie('shareType'))){
			    $query->where('type', cookie('shareType'));
			}
		})
		->order($orderType)
		->count();
		$this->pageTotal = ceil($this->dataTotal/$pageSize);
		$this->listData = $this->pageData->all();
		$userCache=[];
		$userCacheList=[];
		foreach ($this->listData as $key => $value) {
			if(in_array($value["owner"], $userCacheList)){
				$this->listData[$key]["user"] = $userCache[$value["owner"]];
			}else{
				$this->listData[$key]["user"] = Db::name("users")->where("id",$value["owner"])->find();
				array_push($userCacheList,$value["owner"]);
				$userCache[$value["owner"]] = $this->listData[$key]["user"];
			}
		}
		$this->pageNow = input("?get.page")?input("get.page"):1;
	}

	public function listPolicy(){
		$pageSize =!cookie('?pageSize') ? 10 : cookie('pageSize');
		$this->pageData = Db::name("policy")
		->where(function ($query) {
			if(!empty(cookie('policySearch'))){
			    $query->where('policy_name', "like","%".cookie('policySearch')."%");
			}
		})
		->where(function ($query) {
			if(!empty(cookie('policyType'))){
			    $query->where('policy_type', cookie('policyType'));
			}
		})
		->order("id DESC")
		->paginate($pageSize);
		$this->dataTotal = Db::name("policy")
		->where(function ($query) {
			if(!empty(cookie('policySearch'))){
			    $query->where('policy_name', "like","%".cookie('policySearch')."%");
			}
		})
		->where(function ($query) {
			if(!empty(cookie('policyType'))){
			    $query->where('policy_type', cookie('policyType'));
			}
		})
		->order("id DESC")
		->count();
		$this->pageTotal = ceil($this->dataTotal/$pageSize);
		$this->listData = $this->pageData->all();
		$this->pageNow = input("?get.page")?input("get.page"):1;
		foreach ($this->listData as $key => $value) {
			$this->listData[$key]["file_num"] = Db::name("files")->where("policy_id",$value["id"])->count();
			$this->listData[$key]["file_size"] = Db::name("files")->where("policy_id",$value["id"])->sum("size");
		}
	}

	public function getFileInfo($id){
		$fileRecord = Db::name("files")->where("id",$id)->find();
		$policyRecord = Db::name("policy")->where("id",$fileRecord["policy_id"])->find();
		$fileRecord["policy"] = $policyRecord;
		return $fileRecord;
	}

	public function saveThemeFile($options){
		$fileName=$options["name"];
		$dir = ROOT_PATH."application/index/view/";
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
		file_put_contents($dir.rtrim($parentPath,"/")."/".$fileName.".html",$options["content"]);
		return ["error"=>false,"msg"=>"成功"];
	}

	public function saveUser($options){
		if(empty($options["user_pass"])){
			unset($options["user_pass"]);
		}else{
			$options["user_pass"] = md5(config('salt').$options["user_pass"]);
		}
		$userId = $options["uid"];
		unset($options["uid"]);
		try {
			Db::name("users")->where("id",$userId)->update($options);
		} catch (Exception $e) {
			return ["error"=>1,"msg"=>$e->getMessage()];
		}
		return ["error"=>200,"msg"=>"设置已保存"];
	}

	public function banUser($id,$uid){
		if($id == $uid){
			return ["error"=>1,"msg"=>"我的老伙计，你怎么能封禁你自己？"];
		}
		$userData = Db::name("users")->where("id",$id)->find();
		$statusNew = $userData["user_status"] == 1 ? 0 : 1;
		Db::name("users")->where("id",$id)->update(["user_status" => $statusNew]);
		return ["error"=>200,"msg"=>"设置已保存"];
	}

	public function listGroup(){
		$pageSize = empty(cookie('pageSize')) ? 10 : cookie('pageSize');
		$this->pageData = Db::name("groups")
		->order("id DESC")
		->paginate($pageSize);
		$this->dataTotal = Db::name("groups")
		->order("id DESC")
		->count();
		$this->pageTotal = ceil($this->dataTotal/$pageSize);
		$this->listData = $this->pageData->all();
		$this->pageNow = input("?get.page")?input("get.page"):1;
		foreach ($this->listData as $key => $value) {
			$this->listData[$key]["policy"] = Db::name("policy")->where("id",$value["policy_name"])->find();
			$this->listData[$key]["user_num"] = Db::name("users")->where("user_group",$value["id"])->count();
		}
	}
	
	public function addUser($options){
		$options["user_pass"] = md5(config('salt').$options["user_pass"]);
		if(Db::name('users')->where('user_email',$options["user_email"])->find() !=null){
			return ["error" => true,"msg"=>"该邮箱已被注册"];
		}
		$sqlData = [
			'user_email' => $options["user_email"],
			'user_pass' => $options["user_pass"],
			'user_status' => $options["user_status"],
			'user_group' => $options["user_group"],
			'group_primary' => $options["user_group"],
			'user_date' => date("Y-m-d H:i:s"),
			'user_nick' => $options["user_nick"],
			'user_activation_key' => "n",
			'used_storage' => 0,
			'two_step'=>"0",
			'webdav_key' =>$options["user_pass"],
			'delay_time' =>0,
			'avatar' => "default",
			'profile' => true,
		];
		if(Db::name('users')->insert($sqlData)){
			$userId = Db::name('users')->getLastInsID();
			Db::name('folders')->insert( [
				'folder_name' => '根目录',
				'parent_folder' => 0,
				'position' => '.',
				'owner' => $userId,
				'date' => date("Y-m-d H:i:s"),
				'position_absolute' => '/',
			]);
		}
		return ["error"=>0,"msg"=>"设置已保存"];
	}
	
	public function updateOnedriveToken($policyId){
		$policyData = Db::name("policy")->where("id",$policyId)->find();

		if(empty($policyData)){
			throw new \think\Exception("Policy not found");
		}
		$onedrive = new Client([
			'client_id' => $policyData["bucketname"],
		]);
		$url = $onedrive->getLogInUrl([
			'offline_access',
			'files.readwrite.all',
		], Option::getValue("siteURL")."Admin/oneDriveCalllback");
		echo "<script>location.href='".$url."'</script>正在跳转至Onedrive账号授权页面，如果没有跳转，请<a href='$url'>点击这里</a>。";

		Db::name("policy")->where("id",$policyId)->update([
			"sk" => json_encode($onedrive->getState()),
		]);
		\think\Session::set('onedrive.pid',$policyId);
		
	}

	public function oneDriveCalllback($code){
		if(input("?get.error")){
			throw new \think\Exception(input("get.error_description"));
		}
		$policyId = \think\Session::get('onedrive.pid');
		$policyData = Db::name("policy")->where("id",$policyId)->find();
		$onedrive = new Client([
			'client_id' => $policyData["bucketname"],
		
			// Restore the previous state while instantiating this client to proceed in
			// obtaining an access token.
			'state' => json_decode($policyData["sk"]),
		]);
		
		// Obtain the token using the code received by the OneDrive API.
		$onedrive->obtainAccessToken($policyData["ak"], $_GET['code']);
		
		// Persist the OneDrive client' state for next API requests.
		Db::name("policy")->where("id",$policyId)->update([
			"sk" => json_encode($onedrive->getState()),
		]);
		echo "<script>location.href='/Admin/PolicyList?page=1'</script>";
	}

}
?>