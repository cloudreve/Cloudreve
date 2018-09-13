<?php
namespace app\index\model;

require_once 'extend/Qiniu/functions.php';

use think\Model;
use think\Db;
use Qiniu\Auth;
use \app\index\model\Option;
use \app\index\model\FileManage;
use \app\index\model\Task;

class UploadHandler extends Model{

	public $policyId;
	public $policyContent;
	public $userId;
	public $chunkData;
	public $fileSizeTmp;
	public $ossToken;
	public $ossSign;
	public $ossCallback;
	public $ossAccessId;
	public $ossFileName;
	public $ossCallBack;
	public $upyunPolicy;
	public $s3Policy;
	public $s3Sign;
	public $dirName;
	public $s3Credential;
	public $siteUrl;
	public $callBackKey;

	public function __construct($id,$uid){
		$this->policyId = $id;
		$this->userId = $uid;
		$this->policyContent = Db::name('policy')->where('id',$id)->find();
	}

	public function setChunk($chunkId,$chunkSum,$file){
		$sqlData = [
		'user' => $this->userId,
		'ctx' => self::getRandomKey(),
		'obj_name' => self::getRandomKey(8),
		'time' => date("Y-m-d H:i:s"),
		'chunk_id' => $chunkId,
		'sum' => $chunkSum,
		];
		$this->chunkData = $sqlData;
		Db::name('chunks')->insert($sqlData);
		$this->saveChunk($file);
		$this->chunkInfo();
	}

	public function saveChunk($file){
		$chunkSize = strlen($file);
		if(!FileManage::sotrageCheck($this->userId,$chunkSize)){
			$this->setError("空间容量不足",false);
		}
		FileManage::storageCheckOut($this->userId,$chunkSize);
		if($chunkSize >=4195304){
			$this->setError("分片错误",false);
		}
		$chunkObj=fopen (ROOT_PATH . 'public/uploads/chunks/'.$this->chunkData["obj_name"].".chunk","w+");
		$chunkObjWrite = fwrite ($chunkObj,$file);
		if(!$chunkObj || !$chunkObjWrite){
			$this->setError("分片创建错误",false);
		}
	}

	public function chunkInfo(){
		$returnJson = array(
			"ctx" => $this->chunkData["ctx"],
		);
		echo json_encode($returnJson);
		return 0;
	}

	/**
	 * 组合分片并生成最终文件
	 *
	 * @param array $ctx    文件片校验码
	 * @param string $fname 最终文件名
	 * @param string $path  储存目录
	 * @return void
	 */
	public function generateFile($ctx,$fname,$path){
		$ctxTmp = explode(",",$ctx);
		$chunks = Db::name('chunks')->where([
		'ctx' => ["in",$ctxTmp],
		])->order('id asc')->select();
		$file = null;
		if($this->policyContent["policy_type"] != "onedrive"){
			$file = $this->combineChunks($chunks);
		}
		
		$this->filterCheck($file,$fname,$chunks);
		$suffixTmp = explode('.', $fname);
		$fileSuffix = array_pop($suffixTmp);
		if($this->policyContent['autoname']){
			$fileName = $this->getObjName($this->policyContent['namerule'],"local",$fname).".".$fileSuffix;
		}else{
			$fileName = $fname;
		}
		$generatePath = $this->getDirName($this->policyContent['dirrule']);
		$savePath = ROOT_PATH . 'public/uploads/'.$generatePath;
		is_dir($savePath)? :mkdir($savePath,0777,true);
		if(file_exists($savePath.DS.$fileName)){
			$this->setError("文件重名",true,$file,ROOT_PATH . 'public/uploads/chunks/');
		}
		if($this->policyContent["policy_type"] == "onedrive"){
			if($path == "ROOTDIR"){
				$path = "";
			}
			$task = new Task();
			$task->taskName = "Upload Big File " .  $fname . " to Onedrive";
			$task->taskType = "uploadChunksToOnedrive";
			$task->taskContent = json_encode([
				"path" => $path, 
				"fname" => $fname,
				"objname" => $fileName,
				"savePath" =>  $generatePath,
				"fsize" => $this->fileSizeTmp,
				"picInfo" => "",
				"chunks" => $chunks,
				"policyId" => $this->policyContent['id']
			]);
			$task->userId = $this->userId;
			$task->saveTask();
			echo json_encode(array("key" => $fname));
		}else{
			if(!@rename(ROOT_PATH . 'public/uploads/chunks/'.$file,$savePath.DS.$fileName)){
				$this->setError("文件创建失败",true,$file,ROOT_PATH . 'public/uploads/chunks/');
			}else{
				if($path == "ROOTDIR"){
					$path = "";
				}
				$jsonData = array(
					"path" => $path, 
					"fname" => $fname,
					"objname" => $generatePath."/".$fileName,
					"fsize" => $this->fileSizeTmp,
				);
				@list($width, $height, $type, $attr) = getimagesize($savePath.DS.$fileName);
				$picInfo = empty($width)?" ":$width.",".$height;
				$addAction = FileManage::addFile($jsonData,$this->policyContent,$this->userId,$picInfo);
			if(!$addAction[0]){
				$this->setError($addAction[1],true,$fileName,$savePath);
			}
				echo json_encode(array("key" => $fname));
			}
		}
		
	}

	protected function countTotalChunkSize($chunks){
		$size = 0;
		foreach ($chunks as $key => $value) {
			$size += @filesize(ROOT_PATH . 'public/uploads/chunks/'.$value["obj_name"].".chunk");
		}

		return $size;
	}

	public function filterCheck($file,$fname,$chunks){
		if($file !=null){
			$fileSize = filesize(ROOT_PATH . 'public/uploads/chunks/'.$file);
		}else{
			$fileSize = $this->countTotalChunkSize($chunks);
		}
		
		$suffixTmp = explode('.', $fname);
		$fileSuffix = array_pop($suffixTmp);
		$allowedSuffix = explode(',', self::getAllowedExt(json_decode($this->policyContent["filetype"],true)));
		$sufficCheck = !in_array($fileSuffix,$allowedSuffix);
		if(empty(self::getAllowedExt(json_decode($this->policyContent["filetype"],true)))){
			$sufficCheck = false;
		}
		if(($fileSize >= (int)$this->policyContent["max_size"]) || $sufficCheck){
			FileManage::storageGiveBack($this->userId,$fileSize);
			$this->setError("文件效验失败",true,$file,ROOT_PATH . 'public/uploads/chunks/');
		}
		$this->fileSizeTmp = $fileSize;
	}

	/**
	 * 组合文件分片
	 *
	 * @param array $fname 文件分片数据库记录
	 * @return void
	 */
	public function combineChunks($fname){
		$fileName = "file_".self::getRandomKey(8);
		$fileObj=fopen (ROOT_PATH . 'public/uploads/chunks/'.$fileName,"a+");
		$deleteList=[];
		foreach ($fname as $key => $value) {
			$chunkObj = fopen(ROOT_PATH . 'public/uploads/chunks/'.$value["obj_name"].".chunk", "rb");
			if(!$fileObj || !$chunkObj){
				$this->setError("文件创建失败",false);
			}
			$content = fread($chunkObj, 4195304);
			fwrite($fileObj, $content, 4195304);
			unset($content);
			fclose($chunkObj);
			unlink(ROOT_PATH . 'public/uploads/chunks/'.$value["obj_name"].".chunk");
			array_push($deleteList, $value["id"]);
		}
		$chunks = Db::name('chunks')->where([
		'id' => ["in",$deleteList],
		])->delete();
		return $fileName;
	}

	public function fileReceive($file,$info){
		$allowedExt = self::getAllowedExt(json_decode($this->policyContent["filetype"],true));
		$filter = array('size'=>(int)$this->policyContent["max_size"]);
		if(!empty($allowedExt)){
			$filter = array_merge($filter,array("ext" => $allowedExt));
		}
		if(!FileManage::sotrageCheck($this->userId,$file->getInfo('size'))){
			$this->setError("空间容量不足",false);
		}
		if($this->policyContent['autoname']){
			$fileName = $this->getObjName($this->policyContent['namerule'],"local",$file->getInfo('name'));
		}else{
			$fileName = $file->getInfo('name');
		}
		$generatePath = $this->getDirName($this->policyContent['dirrule']);
		$savePath = ROOT_PATH . 'public/uploads/'.$generatePath;
		$Uploadinfo = $file
		->validate($filter)
		->move($savePath,$fileName,false);
		if($Uploadinfo){
			$jsonData = array(
				"path" => $info["path"], 
				"fname" => $info["name"],
				"objname" => $generatePath."/".$Uploadinfo->getSaveName(),
				"fsize" => $Uploadinfo->getSize(),
			);
			@list($width, $height, $type, $attr) = getimagesize(rtrim($savePath, DS).DS.$Uploadinfo->getSaveName());
			$picInfo = empty($width)?" ":$width.",".$height;

			//处理Onedrive等非直传
			if($this->policyContent['policy_type'] == "onedrive"){
				$task = new Task();
				$task->taskName = "Upload" .  $info["name"] . " to Onedrive";
				$task->taskType = "uploadSingleToOnedrive";
				$task->taskContent = json_encode([
					"path" => $info["path"], 
					"fname" => $info["name"],
					"objname" => $Uploadinfo->getSaveName(),
					"savePath" =>  $generatePath,
					"fsize" => $Uploadinfo->getSize(),
					"picInfo" => $picInfo,
					"policyId" => $this->policyContent['id']
				]);
				$task->userId = $this->userId;

				$task->saveTask();

				echo json_encode(array("key" => $info["name"]));
				FileManage::storageCheckOut($this->userId,$jsonData["fsize"],$Uploadinfo->getInfo('size'));
				return;
			}

			//向数据库中添加文件记录
			$addAction = FileManage::addFile($jsonData,$this->policyContent,$this->userId,$picInfo);
			if(!$addAction[0]){
				$tmpFileName = $Uploadinfo->getSaveName();
				unset($Uploadinfo);
				$this->setError($addAction[1],true,$tmpFileName,$savePath);
			}

			//扣除容量
			FileManage::storageCheckOut($this->userId,$jsonData["fsize"],$Uploadinfo->getInfo('size'));
			echo json_encode(array("key" => $info["name"]));
		}else{
			header("HTTP/1.1 401 Unauthorized");
			echo json_encode(array("error" => $file->getError()));
		}
	}

	public function setError($text,$delete = false,$fname="",$path=""){
		header("HTTP/1.1 401 Unauthorized");
		if($delete){
			unlink(rtrim($path, DS).DS.$fname);
		}
		die(json_encode(["error"=> $text]));
	}

	static function getAllowedExt($ext){
		$returnValue = "";
		foreach ($ext as $key => $value) {
			$returnValue .= $value["ext"].",";
		}
		return rtrim($returnValue, ",");
	}

	public function getToken(){
		switch ($this->policyContent['policy_type']) {
			case 'qiniu':
				return $this->getQiniuToken();
				break;
			case 'local':
				return $this->getLocalToken();
				break;
			case 'onedrive':
				return 'nazGTT91tboaLWBC549$:tHSsNyTBxoV4HDfELJeKH1EUmEY=:eyJjYWxsYmFja0JvZHkiOiJ7XCJwYXRoXCI6XCJcIn0iLCJjYWxsYmFja0JvZHlUeXBlIjoiYXBwbGljYXRpb25cL2pzb24iLCJzY29wZSI6ImMxNjMyMTc3LTQ4NGEtNGU1OS1hZDBhLWUwNDc4ZjZhY2NjZSIsImRlYWRsaW5lIjoxNTM2ODMxOTEwfQ==';
				break;
			case 'oss':
				return $this->getOssToken();
				break;
			case 'upyun':
				return $this->getUpyunToken();
				break;
			case 's3':
				return $this->getS3Token();
				break;
			case 'remote':
				return $this->getRemoteToken();
				break;
			default:
				# code...
				break;
		}
	}

	public function getObjName($expression,$type = "qiniu",$origin = ""){
		$policy = array(
			'{date}' =>date("Ymd"),
			'{datetime}' =>date("YmdHis"), 
			'{uid}' =>$this->userId,
			'{timestamp}' =>time(),
			'{randomkey16}' =>self::getRandomKey(16),
			'{randomkey8}' =>self::getRandomKey(8),
			);
		if($type == "qiniu"){
			$policy = array_merge($policy,array("{originname}" => "$(fname)"));
		}else if($type == "local"){
			$policy = array_merge($policy,array("{originname}" => $origin));
		}else if ($type="oss"){
			$policy = array_merge($policy,array("{originname}" => '${filename}'));
		}else if ($type="upyun"){
			$policy = array_merge($policy,array("{originname}" => '{filename}{.suffix}'));
		}
		return strtr($expression,$policy);
	}

	public function getDirName($expression){
		$policy = array(
			'{date}' =>date("Ymd"),
			'{datetime}' =>date("YmdHis"), 
			'{uid}' =>$this->userId,
			'{timestamp}' =>time(),
			'{randomkey16}' =>self::getRandomKey(16),
			'{randomkey8}' =>self::getRandomKey(8),
			);
		return trim(strtr($expression,$policy),"/");
	}

	public function getQiniuToken(){
		$callbackKey = $this->getRandomKey();
		$sqlData = [
		'callback_key' => $callbackKey,
		'pid' => $this->policyId,
		'uid' => $this->userId
		];
		Db::name('callback')->insert($sqlData);
		$auth = new Auth($this->policyContent['ak'], $this->policyContent['sk']);
		$policy = array(
			'callbackUrl' =>Option::getValue("siteURL").'Callback/Qiniu',
			'callbackBody' => '{"fname":"$(fname)","objname":"$(key)","fsize":"$(fsize)","callbackkey":"'.$callbackKey.'","path":"$(x:path)","picinfo":"$(imageInfo.width),$(imageInfo.height)"}',
			'callbackBodyType' => 'application/json',
			'fsizeLimit' => (int)$this->policyContent['max_size'],
		);
		$dirName = $this->getObjName($this->policyContent['dirrule']);
		if($this->policyContent["autoname"]){
			$policy = array_merge($policy,array("saveKey" => $dirName.(empty($dirName)?"":"/").$this->getObjName($this->policyContent['namerule'])));
		}else{
			$policy = array_merge($policy,array("saveKey" => $dirName.(empty($dirName)?"":"/")."$(fname)"));
		}
		if(!empty($this->policyContent['mimetype'])){
			$policy = array_merge($policy,array("mimeLimit" => $this->policyContent['mimetype']));
		}
		$token = $auth->uploadToken($this->policyContent['bucketname'], null, 3600, $policy);
		return $token;
	}

	private function getRemoteToken(){
		$callbackKey = $this->getRandomKey();
		$sqlData = [
			'callback_key' => $callbackKey,
			'pid' => $this->policyId,
			'uid' => $this->userId
		];
		Db::name('callback')->insert($sqlData);
		$policy = array(
			'callbackUrl' =>Option::getValue("siteURL").'Callback/Remote',
			'callbackKey' => $callbackKey,
			'callbackBodyType' => 'application/json',
			'fsizeLimit' => (int)$this->policyContent['max_size'],
			'uid' => $this->userId,
		);
		$dirName = $this->getObjName($this->policyContent['dirrule']);
		if($this->policyContent["autoname"]){
			$policy = array_merge($policy,array("saveKey" => $dirName.(empty($dirName)?"":"/").$this->getObjName($this->policyContent['namerule'])));
		}else{
			$policy = array_merge($policy,array("saveKey" => $dirName.(empty($dirName)?"":"/")."$(fname)"));
		}
		if(!empty($this->policyContent['mimetype'])){
			$policy = array_merge($policy,array("mimeLimit" => $this->policyContent['mimetype']));
		}
		$signingKey = hash_hmac("sha256",json_encode($policy),"UPLOAD".$this->policyContent['sk']);
		$token = $signingKey. ":" .base64_encode(json_encode($policy));
		return $token;
	}

	static function upyunSign($key, $secret, $method, $uri, $date, $policy=null, $md5=null){
		$elems = array();
		foreach (array($method, $uri, $date, $policy, $md5) as $v){
			if ($v){
				$elems[] = $v;
			}
		}
		$value = implode('&', $elems);
		$sign = base64_encode(hash_hmac('sha1', $value, $secret, true));
		return 'UPYUN ' . $key . ':' . $sign;
	}


	public function getUpyunToken(){
		$callbackKey = $this->getRandomKey();
		$sqlData = [
		'callback_key' => $callbackKey,
		'pid' => $this->policyId,
		'uid' => $this->userId
		];
		Db::name('callback')->insert($sqlData);
		$options = Option::getValues(["oss","basic"]);
		$dateNow = gmdate('D, d M Y H:i:s \G\M\T');
		$policy=[
			"bucket" => $this->policyContent['bucketname'],
			"expiration" => time()+$options["timeout"],
			"notify-url" => $options["siteURL"]."Callback/Upyun",
			"content-length-range" =>"0,".$this->policyContent['max_size'],
			"date" => $dateNow,
			"ext-param"=>json_encode([
				"path"=>cookie("path"),
				"uid" => $this->userId,
				"pid" => $this->policyId,
				]),
		];
		$allowedExt = self::getAllowedExt(json_decode($this->policyContent["filetype"],true));
		if(!empty($allowedExt)){
			$policy = array_merge($policy,array("allow-file-type" => $allowedExt));
		}
		$dirName = $this->getObjName($this->policyContent['dirrule']);
		$policy = array_merge($policy,array("save-key" => $dirName.(empty($dirName)?"":"/").uniqid()."CLSUFF{filename}{.suffix}"));
		$this->upyunPolicy = base64_encode(json_encode($policy));
		return self::upyunSign($this->policyContent['op_name'], md5($this->policyContent['op_pwd']), "POST", "/".$this->policyContent['bucketname'],$dateNow,$this->upyunPolicy);
	}

	public function ossCallback(){
		$callbackKey = $this->getRandomKey();
		$sqlData = [
			'callback_key' => $callbackKey,
			'pid' => $this->policyId,
			'uid' => $this->userId
		];
		Db::name('callback')->insert($sqlData);
		$returnValue["callbackUrl"] = Option::getValue("siteUrl").'Callback/Oss';
		$returnValue["callbackBody"] = '{"fname":"${x:fname}","objname":"${object}","fsize":"${size}","callbackkey":"'.$callbackKey.'","path":"${x:path}","picinfo":"${imageInfo.width},${imageInfo.height}"}';
		$this->ossCallBack = base64_encode(json_encode($returnValue));
		return base64_encode(json_encode($returnValue));
	}

	public function getS3Token(){
		$dirName = $this->getDirName($this->policyContent['dirrule']);
		$longDate = gmdate('Ymd\THis\Z');
		$shortDate = gmdate('Ymd');
		$credential = $this->policyContent['ak'] . '/' . $shortDate . '/' . $this->policyContent['op_name'] . '/s3/aws4_request';
		$callbackKey = $this->getRandomKey();
		$sqlData = [
			'callback_key' => $callbackKey,
			'pid' => $this->policyId,
			'uid' => $this->userId
		];
		Db::name('callback')->insert($sqlData);
		$this->siteUrl = Option::getValue("siteUrl");
		$returnValue = [
			"expiration" => date("Y-m-d",time()+1800)."T".date("H:i:s",time()+1800).".000Z",
			"conditions" => [
				0 => ["bucket" => $this->policyContent['bucketname']],
				1 => ["starts-with",'$key', $dirName],
				2 => ["starts-with",'$success_action_redirect' ,$this->siteUrl."Callback/S3/key/".$callbackKey],
				3 => ["content-length-range",1,(int)$this->policyContent['max_size']],
				4 => ['x-amz-algorithm' => 'AWS4-HMAC-SHA256'],
				5 => ['x-amz-credential' => $credential],
				6 => ['x-amz-date' => $longDate],
				7 => ["starts-with", '$name', ""],
				8 => ["starts-with", '$Content-Type', ""],
			]
		];
		$this->s3Policy = base64_encode(json_encode($returnValue));
		$signingKey = hash_hmac("sha256",$shortDate,"AWS4".$this->policyContent['sk'],true);
		$signingKey = hash_hmac("sha256",$this->policyContent['op_name'],$signingKey,true);
		$signingKey = hash_hmac("sha256","s3",$signingKey,true);
		$signingKey = hash_hmac("sha256","aws4_request",$signingKey,true);
		$signingKey = hash_hmac("sha256",$this->s3Policy,$signingKey);
		$this->s3Sign = $signingKey;
		$this->dirName = $dirName;
		$this->s3Credential = $credential;
		$this->x_amz_date = $longDate;
		$this->callBackKey = $callbackKey;
	}

	public function getOssToken(){
		$dirName = $this->getObjName($this->policyContent['dirrule']);
		$returnValu["expiration"] = date("Y-m-d",time()+1800)."T".date("H:i:s",time()+1800).".000Z";
		$returnValu["conditions"][0]["bucket"] = $this->policyContent['bucketname'];
		$returnValu["conditions"][1][0]="starts-with";
		$returnValu["conditions"][1][1]='$key';
		if($this->policyContent["autoname"]){
			$this->ossFileName = $dirName.(empty($dirName)?"":"/").$this->getObjName($this->policyContent['namerule'],"oss");;
		}else{
			$this->ossFileName = $dirName.(empty($dirName)?"":"/").'${filename}';
		}
		$returnValu["conditions"][1][2]=$dirName.(empty($dirName)?"":"/");
		$returnValu["conditions"][2]=["content-length-range",1,(int)$this->policyContent['max_size']];
		$returnValu["conditions"][3]["callback"] = $this->ossCallback();
		$this->ossToken=base64_encode(json_encode($returnValu));
		$this->ossSignToken();
		$this->ossAccessId = $this->policyContent['ak'];
		return false;
	}

	public function ossSignToken(){
		$this->ossSign = base64_encode(hash_hmac("sha1", $this->ossToken, $this->policyContent['sk'],true));  
	}

	public function getLocalToken(){
		$auth = new Auth($this->policyContent['ak'], $this->policyContent['sk']);
		$policy = array(
				'callbackBody' => '{"path":"'.cookie('path').'"}',
				'callbackBodyType' => 'application/json',
		);
		$token = $auth->uploadToken($this->policyContent['bucketname'], null, 3600, $policy);
		return $token;
	}

	static function getRandomKey($length = 16){
		$charTable = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
		$result = ""; 
		for ( $i = 0; $i < $length; $i++ ){ 
			$result .= $charTable[ mt_rand(0, strlen($charTable) - 1) ]; 
		} 
		return $result; 
	}

	static function b64Decode($string) {
		$data = str_replace(array('-','_'),array('+','/'),$string);
		$mod4 = strlen($data) % 4;
		if ($mod4) {
			$data .= substr('====', $mod4);
		}
		return base64_decode($data);
	}

}


?>