<?php
namespace app\index\controller;

use think\Controller;
use think\Db;
use think\Request;
use \app\index\model\Option;
use \app\index\model\User;
use \app\index\model\FileManage;
use \app\index\model\ShareHandler;
use think\Session;


class File extends Controller{

	public $userObj;

	/**
	 * [_initialize description]
	 * @return [type] [description]
	 */
	public function _initialize(){
		$this->userObj = new User(cookie('user_id'),cookie('login_key'));
		if(!$this->userObj->loginStatus){
			echo "Bad request";
			exit();
		}
	}

	public function index(){
		return "";
	}
	 
	public function ListFile(){
		$reqPath = stripslashes(json_decode(file_get_contents("php://input"),true)['path']);
		return json(FileManage::ListFile($reqPath,$this->userObj->uid));
	}

	public function SearchFile(){
		$keyWords = stripslashes(json_decode(file_get_contents("php://input"),true)['path']);
		return json(FileManage::searchFile($keyWords,$this->userObj->uid));
	}

	public function Delete(){
		$reqData = json_decode(file_get_contents("php://input"),true);
		$reqPath = array_key_exists("dirs",$reqData)?$reqData["items"]:array();
		$dirPath = array_key_exists("dirs",$reqData)?$reqData["dirs"]:array();
		FileManage::DirDeleteHandler($dirPath,$this->userObj->uid);
		return json(FileManage::DeleteHandler($reqPath,$this->userObj->uid));
	}

	public function Move(){
		$reqData = json_decode(file_get_contents("php://input"),true);
		$reqPath = array_key_exists("dirs",$reqData)?$reqData["items"]:array();
		$dirPath = array_key_exists("dirs",$reqData)?$reqData["dirs"]:array();
		$newPath = $reqData['newPath'];
		return FileManage::MoveHandler($reqPath,$dirPath,$newPath,$this->userObj->uid);
	}

	public function Rename(){
		$reqPath = json_decode(file_get_contents("php://input"),true)['item'];
		$newPath = json_decode(file_get_contents("php://input"),true)['newItemPath'];
		return FileManage::RenameHandler($reqPath,$newPath,$this->userObj->uid);
	}

	public function Preview(){
		$reqPath =$_GET["path"];
		$fileObj = new FileManage($reqPath,$this->userObj->uid);
		$Redirect = $fileObj->PreviewHandler();
		if($Redirect[0]){
			$this->redirect($Redirect[1],302);
		}
	}
	
	public function ListPic(){
		$reqPath = $_GET["path"];
		return json(FileManage::listPic($reqPath,$this->userObj->uid));
	}

	public function Download(){
		$reqPath = $_GET["path"];
		$fileObj = new FileManage($reqPath,$this->userObj->uid);
		$FileHandler = $fileObj->Download();
		if($FileHandler[0]){
			$this->redirect($FileHandler[1],302);
		}
	}

	public function Share(){
		$reqData = json_decode(file_get_contents("php://input"),true);
		$reqPath = $reqData['item'];
		$shareType = $reqData['shareType'];
		$sharePwd = $reqData['pwd'];
		ShareHandler::createShare($reqPath,$shareType,$sharePwd,$this->userObj->getSQLData(),$this->userObj->getGroupData());
	}

	public function gerSource(){
		$reqData = json_decode(file_get_contents("php://input"),true);
		$reqPath = $reqData['path'];
		$fileObj = new FileManage($reqPath,$this->userObj->uid);
		$FileHandler = $fileObj->Source();
	}

	public function Content(){
		$reqPath = urldecode(input("get.path"));
		$fileObj = new FileManage($reqPath,$this->userObj->uid);
		$FileHandler = $fileObj->getContent();
	}

	public function Edit(){
		$reqPath = json_decode(file_get_contents("php://input"),true)['item'];
		$fileContent = json_decode(file_get_contents("php://input"),true)['content'];
		$fileObj = new FileManage($reqPath,$this->userObj->uid);
		$FileHandler = $fileObj->saveContent($fileContent);
	}

	public function OssDownload(){
		return view('oss_download', [
			'url'  => urldecode(input("get.url")),
			'name' => urldecode(input("get.name")),
		]);
	}

	public function DocPreview(){
		$filePath = input("get.path");
		$fileObj = new FileManage($filePath,$this->userObj->uid);
		$tmpUrl = $fileObj->signTmpUrl();
		$this->redirect("http://view.officeapps.live.com/op/view.aspx?src=".urlencode($tmpUrl),302);
	}

	public function Thumb(){
		$filePath = input("get.path");
		if(input("get.isImg") != "true"){
			return "";
		}
		$fileObj = new FileManage($filePath,$this->userObj->uid);
		$Redirect = $fileObj->getThumb();
		if($Redirect[0]){
			$this->redirect($Redirect[1],302);
		}
	}

	public function GoogleDocPreview(){
		$filePath = input("get.path");
		$fileObj = new FileManage($filePath,$this->userObj->uid);
		$tmpUrl = $fileObj->signTmpUrl();
		$this->redirect("https://docs.google.com/viewer?url=".urlencode($tmpUrl),302);
	}

	/**
	 * [createFolder description]
	 * @Author   Aaron
	 * @DateTime 2017-07-03
	 * @return   [type]     [description]
	 */
	public function createFolder(){
		$reqPath = stripslashes(json_decode(file_get_contents("php://input"),true)['newPath']);
		$pathSplit = explode("/",$reqPath);
		$dirName = end($pathSplit);
		$dirPosition="/";
		foreach ($pathSplit as $key => $value) {
			if (empty($value)){

			}else if($key == (count($pathSplit)-2)){
				$dirPosition = $dirPosition.$value;
			}else if($key == (count($pathSplit)-1)){
			}else{
				$dirPosition = $dirPosition.$value."/";
			}

		}
		return json(FileManage::createFolder($dirName,$dirPosition,$this->userObj->uid));
	} 
}