<?php
namespace app\index\model;

use think\Model;
use think\Db;

use Upyun\Upyun;
use Upyun\Config;

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
		$timeOut = Option::getValue("timeout");
		return [1,\S3\S3::aws_s3_link($this->policyModel["ak"], $this->policyModel["sk"],$this->policyModel["bucketname"],"/".$this->fileModel["pre_name"],3600,$this->policyModel["op_name"])];
	}

	/**
	 * 保存文件内容
	 *
	 * @param string $content 文件内容
	 * @return void
	 */
	public function saveContent($content){
		$s3 = new \S3\S3($this->policyModel["ak"], $this->policyModel["sk"],false,$this->policyModel["op_pwd"]);
		$s3->setSignatureVersion('v4');
		$s3->putObjectString($content, $this->policyModel["bucketname"], $this->fileModel["pre_name"]);
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
		return [1,\S3\S3::aws_s3_link($this->policyModel["ak"], $this->policyModel["sk"],$this->policyModel["bucketname"],"/".$this->fileModel["pre_name"],3600,$this->policyModel["op_name"],array(),false)];
	}
	
	/**
	 * 删除临时文件
	 *
	 * @param string $fname 文件名
	 * @param array $policy 上传策略信息
	 * @return boolean
	 */
	static function deleteS3File($fname,$policy){
		$s3 = new \S3\S3($policy["ak"], $policy["sk"],false,$policy["op_pwd"]);
		$s3->setSignatureVersion('v4');
		return $s3->deleteObject($policy["bucketname"],$fname);
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