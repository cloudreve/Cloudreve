<?php
namespace app\index\model;

use think\Model;
use think\Db;

use \app\index\model\Option;

/**
 * AWS S3策略文件管理适配器
 */
class S3Adapter extends Model{

    private $fileModel;
    private $policyModel;
    private $userModel;

    public function __construct($file,$policy,$user){
        $this->fileModel = $file;
        $this->policyModel = $policy;
        $this->userModel = $user;
    }

	/**
	 * 获取又拍云策略文本文件内容
	 *
	 * @return string 文件内容
	 */
	public function getFileContent(){
		return file_get_contents($this->Preview()[1]);
	}

	/**
	 * 签名S3预览URL
	 *
	 * @return void
	 */
	public function Preview($base=null,$name=null){
		if($base===true || $base ===false){
			$base = null;
		}
		$s3 = new \Aws\S3\S3Client([
			'version' => 'latest',
			'region'  => $this->policyModel["op_name"],
			'endpoint' => $this->policyModel["op_pwd"],
			'use_path_style_endpoint' => true,
			'credentials' => [
					'key'    => $this->policyModel["ak"],
					'secret' => $this->policyModel["sk"],
			],
		]);
		$cmd = $s3->getCommand('GetObject', [
			'Bucket' => $this->policyModel["bucketname"],
			'Key' => $this->fileModel["pre_name"],
		]);
		$timeOut = Option::getValue("timeout");
		$req = $s3->createPresignedRequest($cmd, '+'.($timeOut/60).' minutes');
		$url = (string)$req->getUri();
		
		return [1,$url];
	}

	/**
	 * 保存文件内容
	 *
	 * @param string $content 文件内容
	 * @return void
	 */
	public function saveContent($content){
		$s3 = new \Aws\S3\S3Client([
			'version' => 'latest',
			'region'  => $this->policyModel["op_name"],
			'endpoint' => $this->policyModel["op_pwd"],
			'use_path_style_endpoint' => true,
			'credentials' => [
					'key'    => $this->policyModel["ak"],
					'secret' => $this->policyModel["sk"],
			],
		]);
		$s3->putObject([
			'Bucket' => $this->policyModel["bucketname"],
			'Key' => $this->fileModel["pre_name"],
			'Body' => $content,
		]);
	}

	/**
	 * 删除某一策略下的指定文件
	 *
	 * @param array $fileList   待删除文件的数据库记录
	 * @param array $policyData 待删除文件的上传策略信息
	 * @return void
	 */
	static function DeleteFile($fileList,$policyData){
		foreach (array_column($fileList, 'pre_name') as $key => $value) {
			self::deleteS3File($value,$policyData);
		}
	}

	/**
	 * 生成文件下载URL
	 *
	 * @return array
	 */
	public function Download(){
		$timeOut = Option::getValue("timeout");
		$s3 = new \Aws\S3\S3Client([
			'version' => 'latest',
			'region'  => $this->policyModel["op_name"],
			'endpoint' => $this->policyModel["op_pwd"],
			'use_path_style_endpoint' => true,
			'credentials' => [
					'key'    => $this->policyModel["ak"],
					'secret' => $this->policyModel["sk"],
			],
		]);
		$cmd = $s3->getCommand('GetObject', [
			'Bucket' => $this->policyModel["bucketname"],
			'Key' => $this->fileModel["pre_name"],
			'ResponseContentDisposition' => 'attachment; filename='.$this->fileModel["orign_name"],
		]);
		$timeOut = Option::getValue("timeout");
		$req = $s3->createPresignedRequest($cmd, '+'.($timeOut/60).' minutes');
		$url = (string)$req->getUri();
		
		return [1,$url];
	}
	
	/**
	 * 删除临时文件
	 *
	 * @param string $fname 文件名
	 * @param array $policy 上传策略信息
	 * @return boolean
	 */
	static function deleteS3File($fname,$policy){
		$s3 = new \Aws\S3\S3Client([
			'version' => 'latest',
			'region'  => $policy["op_name"],
			'endpoint' => $policy["op_pwd"],
			'use_path_style_endpoint' => true,
			'credentials' => [
					'key'    =>  $policy["ak"],
					'secret' => $policy["sk"],
			],
		]);
		$result = $s3->deleteObject([
			'Bucket' => $policy["bucketname"],
			'Key' => $fname,
		]);
		return $result["DeleteMarker"];
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