<?php
namespace app\index\model;

use think\Model;
use think\Db;
use Sabre\DAV;
use \app\index\model\FileManage;

class Objects extends DAV\File{

	private $myPath;
	public $id;
	public $fileData;
	public $uid;
	public $fileName;
	public $dir;
	public $actualPath;
	public $isExist;

	function __construct($path) {
		$ex = explode("/",$path);
		$this->uid = $ex[0];
		$t =  strpos($path,$this->uid);
		if($t == 0){
			$this->myPath = substr($path,$t+strlen($this->uid));
		}else{
			$this->myPath = $path;
		}
		$this->fileName = end($ex);
		$this->dir = dirname($this->myPath) == "\\" ? "/" :dirname($this->myPath);
		$this->fileData = Db::name('files')->where('upload_user',$this->uid)->where('dir',$this->dir)->where('orign_name',$this->fileName)->find();
		if(empty($this->fileData)){
			$this->isExist = false;
		}else{
			$this->isExist = true;
		}
		$this->actualPath = ROOT_PATH . 'public/uploads/'.$this->fileData["pre_name"];
	}

	function getName() {
		return $this->fileName;
	}

	function get() {
		return fopen($this->actualPath, 'r');
	}

	function getSize() {
	  return $this->fileData["size"];
	}

	function setName($name){
		$reqPath = $this->myPath;
		$ex = explode("/",$reqPath);
		$newPath = rtrim($reqPath,end($ex)).$name;
		$renameAction = json_decode(FileManage::RenameHandler($reqPath,$newPath,$this->uid,true),true);
		if(!$renameAction["result"]["success"]){
			throw new DAV\Exception\Forbidden($renameAction["result"]["error"]);
		}
	}

	function getContentType() {
		return null;
	}

	function put($data){
		$fileSize = (int)$_SERVER['CONTENT_LENGTH'];
		if(!FileManage::sotrageCheck($this->uid,$fileSize)){
			throw new DAV\Exception\InsufficientStorage("Quota is not enough");
		}
		$filePath = ROOT_PATH . 'public/uploads/' . $this->fileData["pre_name"];
		file_put_contents($filePath, "");
		file_put_contents($filePath, $data);
		FileManage::storageGiveBack($this->uid,$this->fileData["size"]);
		FileManage::storageCheckOut($this->uid,$fileSize);
		@list($width, $height, $type, $attr) = getimagesize($filePath);
		$picInfo = empty($width)?" ":$width.",".$height;
		Db::name('files')->where('id', $this->fileData["id"])->update(['size' => $fileSize,'pic_info' => $picInfo]);
	}

	function delete(){
		FileManage::DeleteHandler([0=>$this->myPath],$this->uid);
	}

	function getETag() {

	    return '"' . sha1(
	        fileinode($this->actualPath) .
	        filesize($this->actualPath) .
	        filemtime($this->actualPath)
	    ) . '"';

	}
}
?>