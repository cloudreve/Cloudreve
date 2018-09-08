<?php
namespace app\index\model;

use think\Model;
use think\Db;

use \app\index\model\Option;

/**
 * 远程策略文件管理适配器
 */
class RemoteAdapter extends Model{

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
	 * 签名文件预览URL
	 *
	 * @return void
	 */
	public function Preview(){
		$remote = new Remote($this->policyModel);
		return [1,$remote->preview($this->fileModel["pre_name"])];
	}

	/**
	 * 保存文件内容
	 *
	 * @param string $content 文件内容
	 * @return bool
	 */
	public function saveContent($content){
		$remote = new Remote($this->policyModel);
		$remote->updateContent($this->fileModel["pre_name"],$content);
	}

	/**
	 * 获取缩略图地址
	 *
	 * @return string 缩略图地址
	 */
	public function getThumb(){
		$remote = new Remote($this->policyModel);
		return [1,$remote->thumb($this->fileModel["pre_name"],explode(",",$this->fileModel["pic_info"]))];
	}

/**
	 * 删除某一策略下的指定文件
	 *
	 * @param array $fileList   待删除文件的数据库记录
	 * @param array $policyData 待删除文件的上传策略信息
	 * @return void
	 */
	static function deleteFile($fileList,$policyData){
		$remoteObj = new Remote($policyData);
		$remoteObj->remove(array_column($fileList, 'pre_name'));
	}

	/**
	 * 生成文件下载URL
	 *
	 * @return array
	 */
	public function Download(){
		$remote = new Remote($this->policyModel);
		return [1,$remote->download($this->fileModel["pre_name"],$this->fileModel["orign_name"])];
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