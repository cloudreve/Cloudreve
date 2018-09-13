<?php
namespace app\index\model;

use think\Model;
use think\Db;
use \app\index\model\Option;
use \app\index\model\LocalAdapter;

class Avatar extends Model{

	public $avatarObj;
	public $errorMsg;
	public $fileName;
	public $avatarKey;
	public $userData;
	public $avatarType;

	public function __construct($new=false,$obj){
		$this->avatarObj = $obj;
		if(!$new){
			$userData = Db::name("users")->where('id',$obj)->find();
			$this->userData = $userData;
			if($userData["avatar"] == "default"){
				$this->avatarType = "default";
			}else{
				$avatarPrarm = explode(".",$userData["avatar"]);
				$this->avatarType = $avatarPrarm[0];
				$this->fileName = ltrim(ltrim($userData["avatar"],"g."),"f.");
			}
		}
	}

	public function SaveAvatar(){
		 $info = $this->avatarObj->validate(['size'=>2097152,'ext'=>'jpg,png,gif.bmp'])->rule('uniqid')->move(ROOT_PATH . 'public' . DS . 'avatars');
		 if($info){
		         $_path = ROOT_PATH.'public/avatars/'.$info->getSaveName();
		         $_img = new Image($_path);
		         $_img->thumb(200, 200);
		         $_img->output();
		         $_img = new Image($_path);
		         $_img->thumb(130, 130);
		         $_img->output("_130");
		         $_img = new Image($_path);
		         $_img->thumb(50, 50);
		         $_img->output("_50");
		         $this->fileName = $info->getSaveName();
		         return true;
		     }else{
		         $this->errorMsg=["result"=>"error","msg"=>$this->avatarObj->getError()];
		         return false;
		     }
	}

	public function bindUser($uid){
		$this->avatarKey = "f.".$this->fileName;
		Db::name("users")->where('id',$uid)->update(["avatar" => $this->avatarKey]);
	}

	public function Out($size){
		switch ($this->avatarType) {
			case 'f':
				$this->outPutFile($size);
				exit();
				break;
			case 'default':
				$this->defaultAvatar($size);
				exit();
				break;
			case 'g':
				return $this->outGravatar($size);
				break;
			default:
				# code...
				break;
		}
	}

	public function outPutFile($size){
		switch ($size) {
			case 's':
				$siezSuffix = "_50";
				break;
			case 'm':
				$siezSuffix = "_130";
				break;
			default:
				$siezSuffix = "";
				break;
		}
		$filePath = ROOT_PATH . 'public/avatars/' . $this->fileName.$siezSuffix;
		if(file_exists($filePath)){
			ob_end_clean();
			header('Content-Type: '.LocalAdapter::getMimetype($filePath)); 
			$fileObj = fopen($filePath,"r");
			while(!feof($fileObj)){
				echo fread($fileObj,2097152);
			}
		}else{
			$this->defaultAvatar($siezSuffix);
		}
	}

	public function outGravatar($size){
		switch ($size) {
			case 's':
				$siezSuffix = "50";
				break;
			case 'm':
				$siezSuffix = "130";
				break;
			default:
				$siezSuffix = "200";
				break;
		}
		ob_end_clean();
		$gravatarServer = Option::getValue("gravatar_server");
		return $gravatarServer.$this->fileName."?d=mm&s=".$siezSuffix;
	}

	public function defaultAvatar($size){
		switch ($size) {
			case 's':
				$siezSuffix = "_50";
				break;
			case 'm':
				$siezSuffix = "_130";
				break;
			default:
				$siezSuffix = "";
				break;
		}
		ob_end_clean();
		$filePath = ROOT_PATH . 'static/img/default.png' .$siezSuffix;
		header('Content-Type: '.LocalAdapter::getMimetype($filePath)); 
			$fileObj = fopen($filePath,"r");
			while(!feof($fileObj)){
				echo fread($fileObj,2097152);
			}
	}

	public function setGravatar(){
		$this->avatarKey="g.".md5($this->userData["user_email"]);
		Db::name("users")->where('id',$this->userData["id"])->update(["avatar" => $this->avatarKey]);
	}

}