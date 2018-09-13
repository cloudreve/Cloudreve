<?php
namespace app\index\model;

use think\Model;
use think\Db;

use OSS\OssClient;
use OSS\Core\OssException;

use \app\index\model\Option;

/**
 * 阿里云OSS策略文件管理适配器
 */
class OssAdapter extends Model{

    private $fileModel;
    private $policyModel;
    private $userModel;

    public function __construct($file,$policy,$user){
        $this->fileModel = $file;
        $this->policyModel = $policy;
        $this->userModel = $user;
    }

	/**
	 * 获取OSS策略文本文件内容
	 *
	 * @return string 文件内容
	 */
	public function getFileContent(){
		return file_get_contents($this->Preview()[1]);
	}

	/**
	 * 签名OSS预览URL
	 *
	 * @return void
	 */
	public function Preview(){
		if(!$this->policyModel['bucket_private']){
			$fileUrl = $this->policyModel["url"].$this->fileModel["pre_name"];
			return[true,$fileUrl];
		}else{
			$accessKeyId = $this->policyModel["ak"];
			$accessKeySecret = $this->policyModel["sk"];
			$endpoint = $this->policyModel["url"];
			try {
				$ossClient = new OssClient($accessKeyId, $accessKeySecret, $endpoint, true);
			} catch (OssException $e) {
				return [false,0];
			}
			$baseUrl = $this->policyModel["url"].$this->fileModel["pre_name"];
			try{
				$signedUrl = $ossClient->signUrl($this->policyModel["bucketname"], $this->fileModel["pre_name"], Option::getValue("timeout"));
			} catch(OssException $e) {
				return [false,0];
			}
			return[true,$signedUrl];
		}
	}

	/**
	 * 保存OSS文件内容
	 *
	 * @param string $content 文件内容
	 * @return void
	 */
	public function saveContent($content){
		$accessKeyId = $this->policyModel["ak"];
		$accessKeySecret = $this->policyModel["sk"];
		$endpoint = "http".ltrim(ltrim($this->policyModel["server"],"https"),"http");
		try {
			$ossClient = new OssClient($accessKeyId, $accessKeySecret, $endpoint, true);
		} catch (OssException $e) {
			die('{ "result": { "success": false, "error": "鉴权失败" } }');
		}
		try{
			$ossClient->putObject($this->policyModel["bucketname"], $this->fileModel["pre_name"], $content);
		} catch(OssException $e) {
			die('{ "result": { "success": false, "error": "编辑失败" } }');
		}
	}

	/**
	 * 获取缩略图地址
	 *
	 * @return string 缩略图地址
	 */
	public function getThumb(){
		if(!$this->policyModel['bucket_private']){
			$fileUrl = $this->policyModel["url"].$this->fileModel["pre_name"]."?x-oss-process=image/resize,m_lfit,h_39,w_90";
			return[true,$fileUrl];
		}else{
			$accessKeyId = $this->policyModel["ak"];
			$accessKeySecret = $this->policyModel["sk"];
			$endpoint = $this->policyModel["url"];
			try {
				$ossClient = new OssClient($accessKeyId, $accessKeySecret, $endpoint, true);
			} catch (OssException $e) {
				return [false,0];
			}
			$baseUrl = $this->policyModel["url"].$this->fileModel["pre_name"];
			try{
				$signedUrl = $ossClient->signUrl($this->policyModel["bucketname"], $this->fileModel["pre_name"], Option::getValue("timeout"),'GET', array("x-oss-process" => 'image/resize,m_lfit,h_39,w_90'));
			} catch(OssException $e) {
				return [false,0];
			}
			return[true,$signedUrl];
		}
	}
	/**
	 * 删除某一策略下的指定OSS文件
	 *
	 * @param array $fileList   待删除文件的数据库记录
	 * @param array $policyData 待删除文件的上传策略信息
	 * @return void
	 */
	static function DeleteFile($fileList,$policyData){
		$accessKeyId = $policyData["ak"];
		$accessKeySecret = $policyData["sk"];
		$endpoint = "http".ltrim(ltrim($policyData["server"],"https"),"http");
		try {
			$ossClient = new OssClient($accessKeyId, $accessKeySecret, $endpoint, true);
		} catch (OssException $e) {
			return false;
		}
		try{
			$ossClient->deleteObjects($policyData["bucketname"], array_column($fileList, 'pre_name'));
		} catch(OssException $e) {
			return false;
		}
	}

	/**
	 * 生成文件下载URL
	 *
	 * @return array
	 */
	public function Download(){
		if(!$this->policyModel['bucket_private']){
			return[true,"/File/OssDownload?url=".urlencode($this->policyModel["url"].$this->fileModel["pre_name"])."&name=".urlencode($this->fileModel["orign_name"])];
		}else{
			$accessKeyId = $this->policyModel["ak"];
			$accessKeySecret = $this->policyModel["sk"];
			$endpoint = $this->policyModel["url"];
			try {
				$ossClient = new OssClient($accessKeyId, $accessKeySecret, $endpoint, true);
			} catch (OssException $e) {
				return [false,0];
			}
			$baseUrl = $this->policyModel["url"].$this->fileModel["pre_name"];
			try{
				$signedUrl = $ossClient->signUrl($this->policyModel["bucketname"], $this->fileModel["pre_name"], Option::getValue("timeout"),'GET', array("response-content-disposition" => 'attachment; filename='.$this->fileModel["orign_name"]));
			} catch(OssException $e) {
				return [false,0];
			}
			return[true,$signedUrl];
		}
	}
	
	/**
	 * 删除临时文件
	 *
	 * @param string $fname 文件名
	 * @param array $policy 上传策略信息
	 * @return boolean
	 */
	static function deleteOssFile($fname,$policy){
		$accessKeyId = $policy["ak"];
		$accessKeySecret = $policy["sk"];
		$endpoint = "http".ltrim(ltrim($policy["server"],"https"),"http");
		try {
			$ossClient = new OssClient($accessKeyId, $accessKeySecret, $endpoint, true);
		} catch (OssException $e) {
			return false;
		}
		try{
			$ossClient->deleteObject($policy["bucketname"], $fname);
		} catch(OssException $e) {
			return false;
		}
		return true;
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