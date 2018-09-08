<?php
namespace app\index\model;

use think\Model;
use think\Db;

use Upyun\Upyun;
use Upyun\Config;

use \app\index\model\Option;

/**
 * 又拍云策略文件管理适配器
 */
class UpyunAdapter extends Model{

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
	 * 签名又拍云预览URL
	 *
	 * @return void
	 */
	public function Preview($base=null,$name=null){
		if($base===true || $base===false){
			$base =null;
		}
		if(!$this->policyModel['bucket_private']){
			$fileUrl = $this->policyModel["url"].$this->fileModel["pre_name"]."?auth=0";
			if(!empty($base)){
				$fileUrl = $base;
			}
			return[true,$fileUrl];
		}else{
			$baseUrl = $this->policyModel["url"].$this->fileModel["pre_name"];
			if(!empty($base)){
				$baseUrl = $base;
			}
			$etime = time() + Option::getValue("timeout");
			$key = $this->policyModel["sk"];
			$path = "/".$this->fileModel["pre_name"];
			if(!empty($name)){
				$path = "/".$name;
			}
			$sign = substr(md5($key.'&'.$etime.'&'.$path), 12, 8).$etime;
			$signedUrl = $baseUrl."?_upt=".$sign;
			return[true,$signedUrl];
		}
	}

	/**
	 * 保存文件内容
	 *
	 * @param string $content 文件内容
	 * @return void
	 */
	public function saveContent($content){
		$bucketConfig = new Config($this->policyModel["bucketname"], $this->policyModel["op_name"], $this->policyModel["op_pwd"]);
		$client = new Upyun($bucketConfig);
		if(empty($content)){
			$content = " ";
		}
		$res=$client->write($this->fileModel["pre_name"],$content);
	}

	/**
     * 计算缩略图大小
     *
     * @param int $width  原始宽
     * @param int $height 原始高
     * @return array
     */
    static function getThumbSize($width,$height){
		$rate = $width/$height;
		$maxWidth = 90;
		$maxHeight = 39;
		$changeWidth = 39*$rate;
		$changeHeight = 90/$rate;
		if($changeWidth>=$maxWidth){
			return [(int)$changeHeight,90];
		}
		return [39,(int)$changeWidth];
    }
    

	/**
	 * 获取缩略图地址
	 *
	 * @return string 缩略图地址
	 */
	public function getThumb(){
		$picInfo = explode(",",$this->fileModel["pic_info"]);
		$thumbSize = self::getThumbSize($picInfo[0],$picInfo[1]);
		$baseUrl =$this->policyModel["url"].$this->fileModel["pre_name"]."!/fwfh/90x39";
		return [1,$this->Preview($baseUrl,$this->fileModel["pre_name"]."!/fwfh/90x39")[1]];
	}

	/**
	 * 删除某一策略下的指定upyun文件
	 *
	 * @param array $fileList   待删除文件的数据库记录
	 * @param array $policyData 待删除文件的上传策略信息
	 * @return void
	 */
	static function DeleteFile($fileList,$policyData){
		foreach (array_column($fileList, 'pre_name') as $key => $value) {
			self::deleteUpyunFile($value,$policyData);
		}
	}

	/**
	 * 生成文件下载URL
	 *
	 * @return array
	 */
	public function Download(){
		return [true,$this->Preview()[1]."&_upd=".urlencode($this->fileModel["orign_name"])];
	}
	
	/**
	 * 删除临时文件
	 *
	 * @param string $fname 文件名
	 * @param array $policy 上传策略信息
	 * @return boolean
	 */
	static function deleteUpyunFile($fname,$policy){
		$bucketConfig = new Config($policy["bucketname"], $policy["op_name"], $policy["op_pwd"]);
		$client = new Upyun($bucketConfig);
		$res=$client->delete($fname,true);
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