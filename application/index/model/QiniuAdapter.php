<?php
namespace app\index\model;

use think\Model;
use think\Db;

require_once   'extend/Qiniu/functions.php';
use Qiniu\Auth;

use \app\index\model\Option;

/**
 * 七牛策略文件管理适配器
 */
class QiniuAdapter extends Model{

    private $fileModel;
    private $policyModel;
    private $userModel;

    public function __construct($file,$policy,$user){
        $this->fileModel = $file;
        $this->policyModel = $policy;
        $this->userModel = $user;
    }

	/**
	 * 获取文本文件内容
	 *
	 * @return string 文件内容
	 */
    public function getFileContent(){
		return file_get_contents($this->Preview()[1]);
	}

	/**
	 * 签名七牛文件预览URL
	 *
	 * @param string $thumb 缩略图参数
	 * @return void
	 */
	public function Preview($thumb=null){
		if($thumb===true || $thumb===false){
			$thumb =null;
		}
		if(!$this->policyModel['bucket_private']){
			$fileUrl = $this->policyModel["url"].$this->fileModel["pre_name"].$thumb;
			return[true,$fileUrl];
		}else{
			$auth = new Auth($this->policyModel["ak"], $this->policyModel["sk"]);
			$baseUrl = $this->policyModel["url"].$this->fileModel["pre_name"].$thumb;
			$signedUrl = $auth->privateDownloadUrl($baseUrl);
			return[true,$signedUrl];
		}
	}

	/**
	 * 保存七牛文件内容
	 *
	 * @param string $content 文件内容
	 * @return bool
	 */
	public function saveContent($content){
		$auth = new Auth($this->policyModel["ak"], $this->policyModel["sk"]);
		$expires = 3600;
		$keyToOverwrite = $this->fileModel["pre_name"];
		$upToken = $auth->uploadToken($this->policyModel["bucketname"], $keyToOverwrite, $expires, null, true);
		$uploadMgr = new \Qiniu\Storage\UploadManager();
		list($ret, $err) = $uploadMgr->put($upToken, $keyToOverwrite, $content);
		if ($err !== null) {
			die('{ "result": { "success": false, "error": "编辑失败" } }');
		} else {
			return true;
		}
	}

	/**
	 * 获取缩略图地址
	 *
	 * @return string 缩略图地址
	 */
	public function getThumb(){
		return $this->Preview("?imageView2/2/w/90/h/39");
	}

/**
	 * 删除某一策略下的指定七牛文件
	 *
	 * @param array $fileList   待删除文件的数据库记录
	 * @param array $policyData 待删除文件的上传策略信息
	 * @return void
	 */
	static function deleteFile($fileList,$policyData){
		$auth = new Auth($policyData["ak"], $policyData["sk"]);
		$config = new \Qiniu\Config();
		$bucketManager = new \Qiniu\Storage\BucketManager($auth);
		$fileListTemp = array_column($fileList, 'pre_name'); 
		$ops = $bucketManager->buildBatchDelete($policyData["bucketname"], $fileListTemp);
		list($ret, $err) = $bucketManager->batch($ops);
	}

	/**
	 * 生成文件下载URL
	 *
	 * @return array
	 */
	public function Download(){
		if(!$this->policyModel['bucket_private']){
			$fileUrl = $this->policyModel["url"].$this->fileModel["pre_name"]."?attname=".urlencode($this->fileModel["orign_name"]);
			return[true,$fileUrl];
		}else{
			$auth = new Auth($this->policyModel["ak"], $this->policyModel["sk"]);
			$baseUrl = $this->policyModel["url"].$this->fileModel["pre_name"]."?attname=".urlencode($this->fileModel["orign_name"]);
			$signedUrl = $auth->privateDownloadUrl($baseUrl);
			return[true,$signedUrl];
		}
	}
	
	/**
	 * 删除临时文件
	 *
	 * @param string $fname 文件名
	 * @param array $policy 上传策略信息
	 * @return void
	 */
	static function deleteSingle($fname,$policy){
		$auth = new Auth($policy["ak"], $policy["sk"]);
		$config = new \Qiniu\Config();
		$bucketManager = new \Qiniu\Storage\BucketManager($auth);
		$err = $bucketManager->delete($policy["bucketname"], $fname);
		if ($err) {
			return false;
		}else{
			return true;
		}
	}

	/**
	 * 签名临时URL用于Office预览
	 *
	 * @return array
	 */
	public function signTmpUrl(){
		return $this->Preview();
	}

}

?>