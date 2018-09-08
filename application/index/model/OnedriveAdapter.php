<?php
namespace app\index\model;

use think\Model;
use think\Db;

use \Krizalys\Onedrive\Client;

use \app\index\model\Option;

/**
 * Onedrive策略文件管理适配器
 */
class OnedriveAdapter extends Model{

    private $fileModel;
    private $policyModel;
	private $userModel;
    private $clinet;

    public function __construct($file,$policy,$user){
        $this->fileModel = $file;
        $this->policyModel = $policy;
		$this->userModel = $user;
		$this->clinet = new Client([
			'stream_back_end' => \Krizalys\Onedrive\StreamBackEnd::TEMP,
			'client_id' => $this->policyModel["bucketname"],
		
			// Restore the previous state while instantiating this client to proceed in
			// obtaining an access token.
			'state' => json_decode($this->policyModel["sk"]),
		]);
    }

	/**
	 * 获取文本文件内容
	 *
	 * @return string 文件内容
	 */
	public function getFileContent(){
		$file = new \Krizalys\Onedrive\File($this->clinet,"/me/drive/root:/".$this->fileModel["pre_name"].":");
		return $file->fetchContent();
	}

	/**
	 * 签名预览URL
	 *
	 * @return void
	 */
	public function Preview($base=null,$name=null){
		$preview = json_decode(json_encode($this->clinet->apiGet("/me/drive/root:/".rawurlencode($this->fileModel["pre_name"]).":")),true);
		return [1,$preview["@microsoft.graph.downloadUrl"]];
	}

	/**
	 * 保存文件内容
	 *
	 * @param string $content 文件内容
	 * @return void
	 */
	public function saveContent($content){
		$this->clinet->createFile(rawurldecode($this->fileModel["pre_name"]),"/me/drive/root:/",$content);
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
		$thumb = json_decode(json_encode($this->clinet->apiGet("/me/drive/root:/".rawurlencode($this->fileModel["pre_name"]).":/thumbnails")),true);
		return [1,$thumb["value"][0]["small"]["url"]];
	}

	/**
	 * 删除某一策略下的指定upyun文件
	 *
	 * @param array $fileList   待删除文件的数据库记录
	 * @param array $policyData 待删除文件的上传策略信息
	 * @return void
	 */
	static function DeleteFile($fileList,$policyData){
		$clinet = new Client([
			'stream_back_end' => \Krizalys\Onedrive\StreamBackEnd::TEMP,
			'client_id' => $policyData["bucketname"],
		
			// Restore the previous state while instantiating this client to proceed in
			// obtaining an access token.
			'state' => json_decode($policyData["sk"]),
		]);
		foreach (array_column($fileList, 'pre_name') as $key => $value) {
			$clinet->deleteObject("/me/drive/root:/".rawurlencode($value).":");
		}
	}

	/**
	 * 生成文件下载URL
	 *
	 * @return array
	 */
	public function Download(){
		$preview = json_decode(json_encode($this->clinet->apiGet("/me/drive/root:/".rawurlencode($this->fileModel["pre_name"]).":")),true);
		return [1,$preview["@microsoft.graph.downloadUrl"]];
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