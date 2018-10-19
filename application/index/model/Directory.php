<?php
namespace app\index\model;

use think\Model;
use think\Db;
use Sabre\DAV;
use \app\index\model\FileManage;
use \app\index\model\User;
use \app\index\model\UploadHandler;

class Directory extends DAV\Node implements DAV\ICollection, DAV\IQuota{

	private $myPath;
	public $uid;
	public $userObj;

	function __construct($path) {
		$ex = explode("/",$path);
		$this->uid = $ex[0];
		if(empty($this->uid)){
			return false;
		}
		$t =  strpos($path,$this->uid);
		if($t == 0){
			$this->myPath = substr($path,$t+strlen($this->uid));
		}else{
			$this->myPath = $path;
		}
		$this->myPath = empty($this->myPath) ? "/" : $this->myPath;
	}

	function createFile($name, $data = NULL){
		$name = str_replace(" ","",$name);
		$userData = Db::name("users")->where("id",$this->uid)->find();
		$groupData = Db::name("groups")->where("id",$userData["user_group"])->find();
		$policyData = Db::name("policy")->where("id",$groupData["policy_name"])->find();
		$uploadHandle = new UploadHandler($groupData["policy_name"],$this->uid);
		$allowedExt = UploadHandler::getAllowedExt(json_decode($policyData["filetype"],true));
		if(!empty($allowedExt)){
			$ex = explode(".",$name);
			$fileSuffix = end($ex);
			if(!in_array($fileSuffix, explode(",",$allowedExt))){
				throw new DAV\Exception\InvalidResourceType('File type not allowed');
			}
		}
		if($policyData["policy_type"] !="local"){
			throw new DAV\Exception\Forbidden('Poliyc not supported yet');
		}
		$fileSize = fstat($data)["size"];
		if(empty($fileSize)){
			$fileSize = -1;
		}
		if($fileSize>$policyData["max_size"]){
			throw new DAV\Exception\InsufficientStorage('File is to large');
		}
		if(!FileManage::sotrageCheck($this->uid,$fileSize)){
			throw new DAV\Exception\InsufficientStorage('Quota is not enough');
		}
		if($policyData['autoname']){
			$fileName = $uploadHandle->getObjName($policyData['namerule'],"local",$name);
		}else{
			$fileName = $name;
		}
		$generatePath = $uploadHandle->getDirName($policyData['dirrule']);
		$savePath = ROOT_PATH . 'public/uploads/'.$generatePath;
		if(!file_exists($savePath)){
			mkdir($savePath,0777,true);
		}
		file_put_contents($savePath."/".$fileName, $data);
		if($fileSize<=0){
			$fileSize = filesize($savePath."/".$fileName);
		}
		$jsonData = array(
			"path" => str_replace("/",",",ltrim($this->myPath,"/")), 
			"fname" => $name,
			"objname" => $generatePath."/".$fileName,
			"fsize" => $fileSize,
		);
		@list($width, $height, $type, $attr) = getimagesize(rtrim($savePath, DS).DS.$fileName);
		$picInfo = empty($width)?" ":$width.",".$height;
		$addAction = FileManage::addFile($jsonData,$policyData,$this->uid,$picInfo);
		if(!$addAction[0]){
			unlink($savePath."/".$fileName);
			throw new DAV\Exception\Conflict($addAction[1]);
		}
		FileManage::storageCheckOut($this->uid,$jsonData["fsize"]);
		//echo json_encode(array("key" => $info["name"]));
	}

	function getQuotaInfo() {
		$this->userObj = new User($this->uid,"",true);
		$quotaInfo = json_decode($this->userObj->getMemory(true),true);
		return [
			$quotaInfo["used"],
			$quotaInfo["total"]
		];

	}

	function setName($name){
		$reqPath = $this->myPath;
		$ex = explode("/",$reqPath);
		$newPath = rtrim(dirname($reqPath) == "\\" ?"/":dirname($reqPath),"/")."/".$name;
		$renameAction = json_decode(FileManage::RenameHandler($reqPath,$newPath,$this->uid,true),true);
		if(!$renameAction["result"]["success"]){
			throw new DAV\Exception\InvalidResourceType($renameAction["result"]["error"]);
		}
	}

	function getChildren() {
		$children = array();
		$fileList = Db::name('files')->where('upload_user',$this->uid)->where('dir',$this->myPath)->select();
		$dirList = Db::name('folders')->where('owner',$this->uid)->where('position',$this->myPath)->select();
		foreach($fileList as $node) {
			// Ignoring files staring with .
			$children[] = $this->getChildFile($node,false);
		}
		foreach($dirList as $node) {
			// Ignoring files staring with .
			$children[] = $this->getChildDir($node,true);
		}
		return $children;
	}

	function getChildFile($name){
		$path = $this->uid.rtrim($this->myPath,"/") . '/' . $name["orign_name"];
		return new Objects($path);
	}

	function getChildDir($name){
		$path = $this->uid.rtrim($this->myPath,"/") . '/' . $name["folder_name"];
		return new Directory($path);
	}

	function delete(){
		foreach ($this->getChildren() as $child) $child->delete();
		FileManage::DirDeleteHandler([0=>$this->myPath],$this->uid);
	}

	function getChild($name) {
		$name = str_replace(" ","",$name);
		if(!$this->childExists($name)){
			throw new DAV\Exception\NotFound('File with name ' . $name . ' could not be located');
		}
		$path = $this->uid.rtrim($this->myPath,"/") . '/' . $name;
		if($this->findDir(rtrim($this->myPath,"/") . '/' . $name)){
			$returnObj = new Directory($path);
			return $returnObj;
		}else{
			return new Objects($path);
		}
	}

	function childExists($name) {
		$name = str_replace(" ","",$name);
		$fileObj = new Objects($this->uid.rtrim($this->myPath,"/") . '/' . $name);
		if($this->findDir(rtrim($this->myPath,"/") . '/' . $name) || $fileObj->isExist){
			return true;
		}
		return false;
	}

	public function findDir($path){
		if($path == "/"){
			return true;
		}
		$explode = explode("/",$path);
		$dirName = end($explode);
		$rootPath  = rtrim($path,"/".$dirName);
		$rootPath = empty($rootPath) ? "/" : $rootPath;
		$dirData = Db::name('folders')->where('owner',$this->uid)->where('position',dirname($path) == "\\" ?"/":dirname($path))->where("folder_name",getDirName($path))->find();
		if(empty($dirData)){
			return false;
		}
		return true;
	}

	function getName() {
		$explode = explode("/", $this->myPath);
		return end($explode);

	}

	function createDirectory($name) {
		$createAction = FileManage::createFolder($name,$this->myPath,$this->uid);
		if(!$createAction["result"]["success"]){
			die($this->myPath);
			throw new DAV\Exception\InvalidResourceType($createAction["result"]["error"]);
		}
	}
}
?>