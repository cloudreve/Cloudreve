<?php
namespace app\index\model;

use think\Model;
use think\Db;
use think\Validate;
use \app\index\model\Option;

class FileManage extends Model{

	public $filePath;
	public $fileData;
	public $userID;
	public $userData;
	public $policyData;
	public $deleteStatus = true;

	private $adapter;

	/**
	 * construct function
	 *
	 * @param string $path 文件路径/文件ID
	 * @param int $uid 用户ID
	 * @param boolean $byId 是否根据文件ID寻找文件
	 */
	public function __construct($path,$uid,$byId=false){
		if($byId){
			$fileRecord = Db::name('files')->where('id',$path)->find();
			$this->filePath = rtrim($fileRecord["dir"],"/")."/".$fileRecord["orign_name"];
		}else{
			$this->filePath = $path;
			$fileInfo = $this->getFileName($path);
			$fileName = $fileInfo[0];
			$path = $fileInfo[1];
			$fileRecord = Db::name('files')->where('upload_user',$uid)->where('orign_name',$fileName)->where('dir',$path)->find();
		}
		if (empty($fileRecord)){
			die('{ "result": { "success": false, "error": "文件不存在" } }');
		}
		$this->fileData = $fileRecord;
		$this->userID = $uid;
		$this->userData = Db::name('users')->where('id',$uid)->find();
		$this->policyData = Db::name('policy')->where('id',$this->fileData["policy_id"])->find();
		switch ($this->policyData["policy_type"]) {
			case 'local':
				$this->adapter = new \app\index\model\LocalAdapter($this->fileData,$this->policyData,$this->userData);
				break;
			case 'qiniu':
				$this->adapter = new \app\index\model\QiniuAdapter($this->fileData,$this->policyData,$this->userData);
				break;
			case 'oss':
				$this->adapter = new \app\index\model\OssAdapter($this->fileData,$this->policyData,$this->userData);
				break;
			case 'upyun':
				$this->adapter = new \app\index\model\UpyunAdapter($this->fileData,$this->policyData,$this->userData);
				break;
			case 's3':
				$this->adapter = new \app\index\model\S3Adapter($this->fileData,$this->policyData,$this->userData);
				break;
			case 'remote':
				$this->adapter = new \app\index\model\RemoteAdapter($this->fileData,$this->policyData,$this->userData);
				break;
			case 'onedrive':
				$this->adapter = new \app\index\model\OnedriveAdapter($this->fileData,$this->policyData,$this->userData);
				break;
			default:
				# code...
				break;
		}
	}

	/**
	 * 获取文件外链地址
	 *
	 * @return void
	 */
	public function Source(){
		if(!$this->policyData["origin_link"]){
			die('{"url":"此文件不支持获取源文件URL"}');
		}else{
			echo ('{"url":"'.$this->policyData["url"].$this->fileData["pre_name"].'"}');
		}
	}

	/**
	 * 获取可编辑文件内容
	 *
	 * @return void
	 */
	public function getContent(){
		$sizeLimit=(int)Option::getValue("maxEditSize");
		if($this->fileData["size"]>$sizeLimit){
			die('{ "result": { "success": false, "error": "您当前用户组最大可编辑'.$sizeLimit.'字节的文件"} }');
		}else{
			try{
				$fileContent = $this->adapter->getFileContent();
			}catch(\Exception $e){
				die('{ "result": { "success": false, "error": "'.$e->getMessage().'"} }');
			}
			$fileContent = $this->adapter->getFileContent();
			$result["result"] = $fileContent;
			if(empty(json_encode($result))){
				$result["result"] = iconv('gb2312','utf-8',$fileContent);
			}
			echo json_encode($result);
		}
	}

	/**
	 * 保存可编辑文件
	 *
	 * @param string $content 要保存的文件内容
	 * @return void
	 */
	public function saveContent($content){
		$contentSize = strlen($content);
		$originSize = $this->fileData["size"];
		if(!FileManage::sotrageCheck($this->userID,$contentSize)){
			die('{ "result": { "success": false, "error": "空间容量不足" } }');
		}
		$this->adapter->saveContent($content);
		FileManage::storageGiveBack($this->userID,$originSize);
		FileManage::storageCheckOut($this->userID,$contentSize);
		Db::name('files')->where('id', $this->fileData["id"])->update(['size' => $contentSize]);
		echo ('{ "result": { "success": true} }');
	}

	/**
	 * 文件名合法性初步检查
	 *
	 * @param string $value 文件名
	 * @return bool 检查结果
	 */
	static function fileNameValidate($value){
		$validate = new Validate([
			'val'  => 'require|max:250',
			'val' => 'chsDash'
		]);
		$data = [
			'val'  => $value
		];
		if (!$validate->check($data)) {
			return false;
		}
		return true;
	}

	/**
	 * 处理重命名
	 *
	 * @param string $fname    原文件路径
	 * @param string $new      新文件路径
	 * @param int $uid         用户ID
	 * @param boolean $notEcho 过程中是否不直接输出结果
	 * @return mixed
	 */
	static function RenameHandler($fname,$new,$uid,$notEcho = false){
		$folderTmp = $new;
		$originFolder = $fname;
		$new = str_replace("/", "", self::getFileName($new)[0]);
		if(!$notEcho){
			$new = str_replace(" ", "", $new);
		}
		if(!self::fileNameValidate($new)){
			if($notEcho){
				return '{ "result": { "success": false, "error": "文件名只支持数字、字母、下划线" } }';
			}
			die('{ "result": { "success": false, "error": "文件名只支持数字、字母、下划线" } }');
		}
		$path = self::getFileName($fname)[1];
		$fname = self::getFileName($fname)[0];
		$fileRecord = Db::name('files')->where('upload_user',$uid)->where('orign_name',$fname)->where('dir',$path)->find();
		if (empty($new)){
			if($notEcho){
					return '{ "result": { "success": false, "error": "文件重名或文件名非法" } }';
			}
			die('{ "result": { "success": false, "error": "文件重名或文件名非法" } }');
		}
		if(empty($fileRecord)){
			self::folderRename($originFolder,$folderTmp,$uid,$notEcho);
			die();
		}
		$originSuffix = explode(".",$fileRecord["orign_name"]);
		$newSuffix = explode(".",$new);
		if(end($originSuffix) != end($newSuffix)){
			if($notEcho){
					return '{ "result": { "success": false, "error": "请不要更改文件扩展名" } }';
			}
			die('{ "result": { "success": false, "error": "请不要更改文件扩展名" } }');
		}
		Db::name('files')->where([
			'upload_user' => $uid,
			'dir' => $path,
			'orign_name' =>$fname,
		])->setField('orign_name', $new);
		if($notEcho){
				return '{ "result": { "success": true} }';
		}
		echo ('{ "result": { "success": true} }');
	}

	/**
	 * 处理目录重命名
	 *
	 * @param string $fname    原文件路径
	 * @param string $new      新文件路径
	 * @param int $uid         用户ID
	 * @param boolean $notEcho 过程中是否不直接输出结果
	 * @return void
	 */
	static function folderRename($fname,$new,$uid,$notEcho = false){
		$newTmp = $new;
		$nerFolderTmp = explode("/",$new);
		$new = array_pop($nerFolderTmp);
		$oldFolderTmp = explode("/",$fname);
		$old = array_pop($oldFolderTmp);
		if(!self::fileNameValidate($new)){
			if($notEcho){
				return '{ "result": { "success": false, "error": "目录名只支持数字、字母、下划线" } }';
			}
			die('{ "result": { "success": false, "error": "目录名只支持数字、字母、下划线" } }');
		}
		$folderRecord = Db::name('folders')->where('owner',$uid)->where('position_absolute',$fname)->find();
		if(empty($folderRecord)){
			if($notEcho){
				return '{ "result": { "success": false, "error": "目录不存在" } }';
			}
			die('{ "result": { "success": false, "error": "目录不存在" } }');
		}
		$newPositionAbsolute = substr($fname, 0, strrpos( $fname, '/'))."/".$new;
		Db::name('folders')->where('owner',$uid)->where('position_absolute',$fname)->update([
			'folder_name' => $new,
			'position_absolute' => $newPositionAbsolute,
		]);
		$childFolder = Db::name('folders')->where('owner',$uid)->where('position',"like",$fname."%")->select();
		foreach ($childFolder as $key => $value) {
			$tmpPositionAbsolute = "";
			$tmpPosition = "";
			$pos = strpos($value["position_absolute"], $fname);   
			if ($pos === false) {   
				$tmpPositionAbsolute = $value["position_absolute"];   
			}   
			$tmpPositionAbsolute = substr_replace($value["position_absolute"], $newTmp, $pos, strlen($fname));
			$pos = strpos($value["position"], $fname);   
			if ($pos === false) {   
				$tmpPosition = $value["position"];   
			}   
			$tmpPosition = substr_replace($value["position"], $newTmp, $pos, strlen($fname));
			Db::name('folders')->where('id',$value["id"])->update([
				'position_absolute' => $tmpPositionAbsolute,
				'position' =>$tmpPosition,
			]);
		}
		$childFiles = Db::name('files')->where('upload_user',$uid)->where('dir',"like",$fname."%")->select();
		foreach ($childFiles as $key => $value) {
			$tmpPosition = "";
			$pos = strpos($value["dir"], $fname);   
			if ($pos === false) {   
				$tmpPosition = $value["dir"];   
			}   
			$tmpPosition = substr_replace($value["dir"], $newTmp, $pos, strlen($fname));
			Db::name('files')->where('id',$value["id"])->update([
				'dir' =>$tmpPosition,
			]);
		}
		if($notEcho){
				return '{ "result": { "success": true} }';
			}
		echo ('{ "result": { "success": true} }');
	}

	/**
	 * 根据文件路径获取文件名和父目录路径
	 *
	 * @param string 文件路径
	 * @return array 
	 */
	static function getFileName($path){
		$pathSplit = explode("/",$path);
		$fileName = end($pathSplit);
		$pathSplitDelete = array_pop($pathSplit);
		$path="";
		foreach ($pathSplit as $key => $value) {
			if (empty($value)){

			}else{
				$path =$path."/".$value;
			}
		} 
		$path = empty($path)?"/":$path;
		return [$fileName,$path];
	}

	/**
	 * 处理文件预览
	 *
	 * @param boolean $isAdmin 是否为管理员预览
	 * @return array 重定向信息
	 */
	public function PreviewHandler($isAdmin=false){
		return $this->adapter->Preview($isAdmin);
	}

	/**
	 * 获取图像缩略图
	 *
	 * @return array 重定向信息
	 */
	public function getThumb(){
		return $this->adapter->getThumb();
	}

	/**
	 * 处理文件下载
	 *
	 * @param boolean $isAdmin 是否为管理员请求
	 * @return array 文件下载URL
	 */
	public function Download($isAdmin=false){
		return $this->adapter->Download($isAdmin);
	}

	/**
	 * 处理目录删除
	 *
	 * @param string $path 目录路径
	 * @param int $uid     用户ID
	 * @return void
	 */
	static function DirDeleteHandler($path,$uid){
		global $toBeDeleteDir;
		global $toBeDeleteFile;
		$toBeDeleteDir = [];
		$toBeDeleteFile = [];
		foreach ($path as $key => $value) {
			array_push($toBeDeleteDir,$value);
		}
		
		foreach ($path as $key => $value) {
			self::listToBeDelete($value,$uid);
		}
		if(!empty($toBeDeleteFile)){
			self::DeleteHandler($toBeDeleteFile,$uid);
		}
		if(!empty($toBeDeleteDir)){
			self::deleteDir($toBeDeleteDir,$uid);
		}
	}

	/**
	 * 列出待删除文件或目录
	 *
	 * @param string $path 对象路径
	 * @param int $uid     用户ID
	 * @return void
	 */
	static function listToBeDelete($path,$uid){
		global $toBeDeleteDir;
		global $toBeDeleteFile;
		$fileData = Db::name('files')->where([
		'dir' => $path,
		'upload_user' => $uid,
		])->select();
		foreach ($fileData as $key => $value) {
			array_push($toBeDeleteFile,$path."/".$value["orign_name"]);
		}
		$dirData = Db::name('folders')->where([
		'position' => $path,
		'owner' => $uid,
		])->select();
		foreach ($dirData as $key => $value) {
			array_push($toBeDeleteDir,$value["position_absolute"]);
			self::listToBeDelete($value["position_absolute"],$uid);
		}
	}

	/**
	 * 删除目录
	 *
	 * @param string $path 目录路径
	 * @param int $uid     用户ID
	 * @return void
	 */
	static function deleteDir($path,$uid){
		Db::name('folders')
		->where("owner",$uid)
		->where([
		'position_absolute' => ["in",$path],
		])->delete();
	}

	/**
	 * 处理删除请求
	 *
	 * @param string $path 路径
	 * @param int $uid     用户ID
	 * @return array
	 */
	static function DeleteHandler($path,$uid){
		if(empty($path)){
			return ["result"=>["success"=>true,"error"=>null]];
		}
		foreach ($path as $key => $value) {
			$fileInfo = self::getFileName($value);
			$fileName = $fileInfo[0];
			$filePath = $fileInfo[1];
			$fileNames[$key] = $fileName;
			$filePathes[$key] = $filePath;
		}
		$fileData = Db::name('files')->where([
		'orign_name' => ["in",$fileNames],
		'dir' => ["in",$filePathes],
		'upload_user' => $uid,
		])->select();
		$fileListTemp=[];
		$uniquePolicy = self::uniqueArray($fileData);
		foreach ($fileData as $key => $value) {
			if(empty($fileListTemp[$value["policy_id"]])){
				$fileListTemp[$value["policy_id"]] = [];
			}
			array_push($fileListTemp[$value["policy_id"]],$value);
		}
		foreach ($fileListTemp as $key => $value) {
			if(in_array($key,$uniquePolicy["qiniuList"])){
				QiniuAdapter::DeleteFile($value,$uniquePolicy["qiniuPolicyData"][$key][0]);
				self::deleteFileRecord(array_column($value, 'id'),array_sum(array_column($value, 'size')),$value[0]["upload_user"]);
			}else if(in_array($key,$uniquePolicy["localList"])){
				LocalAdapter::DeleteFile($value,$uniquePolicy["localPolicyData"][$key][0]);
				self::deleteFileRecord(array_column($value, 'id'),array_sum(array_column($value, 'size')),$value[0]["upload_user"]);
			}else if(in_array($key,$uniquePolicy["ossList"])){
				OssAdapter::DeleteFile($value,$uniquePolicy["ossPolicyData"][$key][0]);
				self::deleteFileRecord(array_column($value, 'id'),array_sum(array_column($value, 'size')),$value[0]["upload_user"]);
			}else if(in_array($key,$uniquePolicy["upyunList"])){
				UpyunAdapter::DeleteFile($value,$uniquePolicy["upyunPolicyData"][$key][0]);
				self::deleteFileRecord(array_column($value, 'id'),array_sum(array_column($value, 'size')),$value[0]["upload_user"]);
			}else if(in_array($key,$uniquePolicy["s3List"])){
				S3Adapter::DeleteFile($value,$uniquePolicy["s3PolicyData"][$key][0]);
				self::deleteFileRecord(array_column($value, 'id'),array_sum(array_column($value, 'size')),$value[0]["upload_user"]);
			}else if(in_array($key,$uniquePolicy["remoteList"])){
				RemoteAdapter::DeleteFile($value,$uniquePolicy["remotePolicyData"][$key][0]);
				self::deleteFileRecord(array_column($value, 'id'),array_sum(array_column($value, 'size')),$value[0]["upload_user"]);
			}else if(in_array($key,$uniquePolicy["onedriveList"])){
				OnedriveAdapter::DeleteFile($value,$uniquePolicy["onedrivePolicyData"][$key][0]);
				self::deleteFileRecord(array_column($value, 'id'),array_sum(array_column($value, 'size')),$value[0]["upload_user"]);
			}
		}
		return ["result"=>["success"=>true,"error"=>null]];
	}

	/**
	 * 处理移动
	 *
	 * @param array $file 文件路径列表
	 * @param array $dir  目录路径列表
	 * @param string $new 新路径
	 * @param int $uid    用户ID
	 * @return void
	 */
	static function MoveHandler($file,$dir,$new,$uid){
		if(in_array($new,$dir)){
			die('{ "result": { "success": false, "error": "不能移动目录到自身" } }');
		}
		if(Db::name('folders')->where('owner',$uid)->where('position_absolute',$new)->find() == null){
			die('{ "result": { "success": false, "error": "目录不存在" } }');
		}
		$moveName=[];
		$movePath=[];
		foreach ($file as $key => $value) {
			$fileInfo = self::getFileName($value);
			$moveName[$key] = $fileInfo[0];
			$movePath[$key] = $fileInfo[1];
		}
		$dirName=[];
		$dirPa=[];
		foreach ($dir as $key => $value) {
			$dirInfo = self::getFileName($value);
			$dirName[$key] = $dirInfo[0];
			$dirPar[$key] = $dirInfo[1];
		}
		$nameCheck = Db::name('files')->where([
			'upload_user' => $uid,
			'dir' => $new,
			'orign_name' =>["in",$moveName],
		])->find();
		$dirNameCheck = array_merge($dirName,$moveName);
		$dirCheck = Db::name('folders')->where([
			'owner' => $uid,
			'position' => $new,
			'folder_name' =>["in",$dirNameCheck],
		])->find();
		if($nameCheck || $dirCheck){
			die('{ "result": { "success": false, "error": "文件名冲突，请检查是否重名" } }');
		}
		if(!empty($dir)){
			die('{ "result": { "success": false, "error": "暂不支持移动目录" } }');
		}
		Db::name('files')->where([
			'upload_user' => $uid,
			'dir' => ["in",$movePath],
			'orign_name' =>["in",$moveName],
		])->setField('dir', $new);
		echo ('{ "result": { "success": true} }');
	}

	/**
	 * ToDo 移动文件
	 *
	 * @param array $file
	 * @param string $path
	 * @return void
	 */
	static function moveFile($file,$path){

	}

	static function deleteFileRecord($id,$size,$uid){
		Db::name('files')->where([
		'id' => ["in",$id],
		])->delete();
		Db::name('shares')
		->where(['owner' => $uid])
		->where(['source_type' => "file"])
		->where(['source_name' => ["in",$id],])
		->delete();
		Db::name('users')->where([
		'id' => $uid,
		])->setDec('used_storage', $size);
	}

	/**
	 * [List description]
	 * @param [type] $path [description]
	 * @param [type] $uid  [description]
	 */
	static function ListFile($path,$uid){
		$fileList = Db::name('files')->where('upload_user',$uid)->where('dir',$path)->select();
		$dirList = Db::name('folders')->where('owner',$uid)->where('position',$path)->select();
		$count= 0;
		$fileListData=[];
		foreach ($dirList as $key => $value) {
			$fileListData['result'][$count]['name'] = $value['folder_name'];
			$fileListData['result'][$count]['rights'] = "drwxr-xr-x";
			$fileListData['result'][$count]['size'] = "0";
			$fileListData['result'][$count]['date'] = $value['date'];
			$fileListData['result'][$count]['type'] = 'dir';
			$fileListData['result'][$count]['name2'] = "";
			$fileListData['result'][$count]['id'] = $value['id'];
			$fileListData['result'][$count]['pic'] = "";
			$count++;
		}
		foreach ($fileList as $key => $value) {
			$fileListData['result'][$count]['name'] = $value['orign_name'];
			$fileListData['result'][$count]['rights'] = "drwxr-xr-x";
			$fileListData['result'][$count]['size'] = $value['size'];
			$fileListData['result'][$count]['date'] = $value['upload_date'];
			$fileListData['result'][$count]['type'] = 'file';
			$fileListData['result'][$count]['name2'] = $value["dir"];
			$fileListData['result'][$count]['id'] = $value["id"];
			$fileListData['result'][$count]['pic'] = $value["pic_info"];
			$count++;
		}
	
		return $fileListData;
	}

	static function listPic($path,$uid,$url="/File/Preview?"){
		$firstPreview = self::getFileName($path);
		$path=$firstPreview[1];
		$fileList = Db::name('files')
		->where('upload_user',$uid)
		->where('dir',$path)
		->where('pic_info',"<>"," ")
		->where('pic_info',"<>","0,0")
		->where('pic_info',"<>","null,null")
		->select();
		$count= 0;
		$fileListData=[];
		foreach ($fileList as $key => $value) {
			if($value["orign_name"] == $firstPreview[0]){
				$previewPicInfo = explode(",",$value["pic_info"]);
				$previewSrc = $url."action=preview&path=".$path."/".$value["orign_name"];
			}else{
				$picInfo = explode(",",$value["pic_info"]);
				$fileListData[$count]['src'] = $url."action=preview&path=".$path."/".$value["orign_name"];
				$fileListData[$count]['w'] = $picInfo[0];
				$fileListData[$count]['h'] = $picInfo[1];
				$fileListData[$count]['title'] = $value["orign_name"];
				$count++;
			}
		}
		array_unshift($fileListData,array(
			'src' => $previewSrc,
			'w' => $previewPicInfo[0],
			'h' => $previewPicInfo[1],
			'title' => $firstPreview[0],
			));
		return $fileListData;
	}

	/**
	 * [createFolder description]
	 * @param  [type] $dirName     [description]
	 * @param  [type] $dirPosition [description]
	 * @param  [type] $uid         [description]
	 * @return [type]              [description]
	 */
	static function createFolder($dirName,$dirPosition,$uid){
		$dirName = str_replace(" ","",$dirName);
		$dirName = str_replace("/","",$dirName);
		if(empty($dirName)){
			return ["result"=>["success"=>false,"error"=>"目录名不能为空"]];
		}
		if(Db::name('folders')->where('position_absolute',$dirPosition)->where('owner',$uid)->find() ==null || Db::name('folders')->where('owner',$uid)->where('position',$dirPosition)->where('folder_name',$dirName)->find() !=null || Db::name('files')->where('upload_date',$uid)->where('dir',$dirPosition)->where('pre_name',$dirName)->find() !=null){
			return ["result"=>["success"=>false,"error"=>"路径不存在或文件已存在"]];
		}
		$sqlData = [
			'folder_name' => $dirName,
			'parent_folder' => Db::name('folders')->where('position_absolute',$dirPosition)->value('id'),
			'position' => $dirPosition,
			'owner' => $uid,
			'date' => date("Y-m-d H:i:s"),
			'position_absolute' => ($dirPosition == "/")?($dirPosition.$dirName):($dirPosition."/".$dirName),
			];
		if(Db::name('folders')->insert($sqlData)){
			return ["result"=>["success"=>true,"error"=>null]];
		}

	}

	static function getTotalStorage($uid){
		$userData = Db::name('users')->where('id',$uid)->find();
		$basicStronge = Db::name('groups')->where('id',$userData['user_group'])->find();
		$addOnStorage = Db::name('storage_pack')
		->where('uid',$uid)
		->where('dlay_time',">",time())
		->sum('pack_size');
		return $addOnStorage+$basicStronge["max_storage"];
	}

	static function getUsedStorage($uid){
		$userData = Db::name('users')->where('id',$uid)->find();
		return $userData['used_storage'];
	}

	static function sotrageCheck($uid,$fsize){
		$totalStorage = self::getTotalStorage($uid);
		$usedStorage = self::getUsedStorage($uid);
		return ($totalStorage > ($usedStorage + $fsize)) ? True : False;
	}

	static function storageCheckOut($uid,$size){
		Db::name('users')->where('id',$uid)->setInc('used_storage',$size);
	}

	static function storageGiveBack($uid,$size){
		Db::name('users')->where('id',$uid)->setDec('used_storage',$size);
	}

	static function addFile($jsonData,$policyData,$uid,$picInfo=" "){
		$dir = "/".str_replace(",","/",$jsonData['path']);
		$fname = $jsonData['fname'];
		if(self::isExist($dir,$fname,$uid)){
			return[false,"文件已存在"];
		}
		$folderBelong = Db::name('folders')->where('owner',$uid)->where('position_absolute',$dir)->find();
		if($folderBelong ==null){
			return[false,"目录不存在"];
		}
		$sqlData = [
			'orign_name' => $jsonData['fname'],
			'pre_name' => $jsonData['objname'],
			'upload_user' => $uid,
			'size' => $jsonData['fsize'],
			'upload_date' => date("Y-m-d H:i:s"),
			'parent_folder' => $folderBelong['id'],
			'policy_id' => $policyData['id'],
			'dir' => $dir,
			'pic_info' => $picInfo,
		];
		if(Db::name('files')->insert($sqlData)){
			return [true,"上传成功"];
		}

	}

	static function isExist($dir,$fname,$uid){
		if(Db::name('files')->where('upload_user',$uid)->where('dir',$dir)->where('orign_name',$fname)->find() !=null){
			return true;
		}else{
			return false;
		}
	}

	static function deleteFile($fname,$policy){
		switch ($policy['policy_type']) {
			case 'qiniu':
				return QiniuAdapter::deleteSingle($fname,$policy);
				break;
			case 'oss':
				return OssAdapter::deleteOssFile($fname,$policy);
				break;
			case 'upyun':
				return UpyunAdapter::deleteUpyunFile($fname,$policy);
				break;
			case 's3':
				return S3Adapter::deleteS3File($fname,$policy);
				break;
			default:
				# code...
				break;
		}
	}

	static function uniqueArray($data = array()){
		$tempList = [];
		$qiniuList = [];
		$qiniuPolicyData = [];
		$localList = [];
		$localPolicyData = [];
		$ossList = [];
		$ossPolicyData = [];
		$upyunList = [];
		$upyunPolicyData = [];
		$s3List = [];
		$s3PolicyData = [];
		$remoteList = [];
		$remotePolicyData = [];
		$onedriveList = [];
		$onedrivePolicyData = [];
		foreach ($data as $key => $value) {
			if(!in_array($value['policy_id'],$tempList)){
				array_push($tempList,$value['policy_id']);
				$policyTempData = Db::name('policy')->where('id',$value['policy_id'])->find();
				switch ($policyTempData["policy_type"]) {
					case 'qiniu':
						array_push($qiniuList,$value['policy_id']);
						if(empty($qiniuPolicyData[$value['policy_id']])){
							$qiniuPolicyData[$value['policy_id']] = [];
						}
						array_push($qiniuPolicyData[$value['policy_id']],$policyTempData);
						break;
					case 'local':
						array_push($localList,$value['policy_id']);
						if(empty($localPolicyData[$value['policy_id']])){
							$localPolicyData[$value['policy_id']] = [];
						}
						array_push($localPolicyData[$value['policy_id']],$policyTempData);
						break;
					case 'oss':
						array_push($ossList,$value['policy_id']);
						if(empty($ossPolicyData[$value['policy_id']])){
							$ossPolicyData[$value['policy_id']] = [];
						}
						array_push($ossPolicyData[$value['policy_id']],$policyTempData);
						break;
					case 'upyun':
						array_push($upyunList,$value['policy_id']);
						if(empty($upyunPolicyData[$value['policy_id']])){
							$upyunPolicyData[$value['policy_id']] = [];
						}
						array_push($upyunPolicyData[$value['policy_id']],$policyTempData);
						break;
					case 's3':
						array_push($s3List,$value['policy_id']);
						if(empty($s3PolicyData[$value['policy_id']])){
							$s3PolicyData[$value['policy_id']] = [];
						}
						array_push($s3PolicyData[$value['policy_id']],$policyTempData);
						break;
					case 'remote':
						array_push($remoteList,$value['policy_id']);
						if(empty($remotePolicyData[$value['policy_id']])){
							$remotePolicyData[$value['policy_id']] = [];
						}
						array_push($remotePolicyData[$value['policy_id']],$policyTempData);
						break;
					case 'onedrive':
						array_push($onedriveList,$value['policy_id']);
						if(empty($onedrivePolicyData[$value['policy_id']])){
							$onedrivePolicyData[$value['policy_id']] = [];
						}
						array_push($onedrivePolicyData[$value['policy_id']],$policyTempData);
						break;
					default:
						# code...
						break;
				}
			}
		}
		$returenValue=array(
			'policyId' => $tempList ,
			'qiniuList' => $qiniuList,
			'qiniuPolicyData' => $qiniuPolicyData,
			'localList' => $localList,
			'localPolicyData' => $localPolicyData,
			'ossList' => $ossList,
			'ossPolicyData' => $ossPolicyData,
			'upyunList' => $upyunList,
			'upyunPolicyData' => $upyunPolicyData,
			's3List' => $s3List,
			's3PolicyData' => $s3PolicyData,
			'remoteList' => $remoteList,
			'remotePolicyData' => $remotePolicyData,
			'onedriveList' => $onedriveList,
			'onedrivePolicyData' => $onedrivePolicyData,
		);
		return $returenValue;
	}

	public function signTmpUrl(){
		return $this->adapter->signTmpUrl()[1];
	}

}
?>