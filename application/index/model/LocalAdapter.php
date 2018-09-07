<?php
namespace app\index\model;

use think\Model;
use think\Db;

use \app\index\model\Option;

/**
 * 本地策略文件管理适配器
 */
class LocalAdapter extends Model{

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
     * @return string 内容
     */
    public function getFileContent(){
        $filePath = ROOT_PATH . 'public/uploads/' . $this->fileModel["pre_name"];
        $fileObj = fopen($filePath,"r");
		$fileContent = fread($fileObj,filesize($filePath)+1);
		return $fileContent;
    }
    
    /**
	 * 保存可编辑文件
	 *
	 * @param string $content 要保存的文件内容
	 * @return void
	 */
    public function saveContent($content){
        $filePath = ROOT_PATH . 'public/uploads/' . $this->fileModel["pre_name"];
		file_put_contents($filePath, "");
		file_put_contents($filePath, $content);
    }

    /**
     * 输出预览文件
     *
     * @param boolean $isAdmin 是否为管理员请求
     * @return mixed 文件数据
     */
    public function Preview($isAdmin = false){
        $speedLimit = Db::name('groups')->where('id',$this->userModel["user_group"])->find();
		$rangeTransfer = $speedLimit["range_transfer"];
		$speedLimit = $speedLimit["speed"];
		$sendFileOptions = Option::getValues(["download"]);
		if($sendFileOptions["sendfile"] == "1" && !empty($sendFileOptions)){
			$this->sendFile($speedLimit,$rangeTransfer,false,$sendFileOptions["header"]);
		}else{
			if($isAdmin){
				$speedLimit="";
			}
			if($speedLimit == "0"){
				exit();
			}else if(empty($speedLimit)){
				header("Cache-Control: max-age=10800");
				$this->outputWithoutLimit(false,$rangeTransfer);
				exit();
			}else if((int)$speedLimit > 0){
				header("Cache-Control: max-age=10800");
				$this->outputWithLimit($speedLimit);
			}
		}
    }

    /**
     * 使用Sendfile模式发送文件数据
     *
     * @param int $speed        下载限速
     * @param boolean $range    是否支持断点续传
     * @param boolean $download 是否为下载请求
     * @param string $header    Sendfile Header
     * @return void
     */
    private function sendFile($speed,$range,$download=false,$header="X-Sendfile"){
		$filePath = ROOT_PATH . 'public/uploads/' . $this->fileModel["pre_name"];
		$realPath = ROOT_PATH . 'public/uploads/' . $this->fileModel["pre_name"];
		if($header == "X-Accel-Redirect"){
			$filePath = '/public/uploads/' . $this->fileModel["pre_name"];
		}
		if($download){
			$filePath = str_replace("\\","/",$filePath);
			if($header == "X-Accel-Redirect"){
				ob_flush();
				flush();
				echo "s";
			}
			//保证如下顺序，否则最终浏览器中得到的content-type为'text/html'
			//1,写入 X-Sendfile 头信息
			$pathToFile = str_replace('%2F', '/', rawurlencode($filePath));
			header($header.": ".$pathToFile);
			//2,写入Content-Type头信息
			$mime_type = self::getMimetypeOnly($realPath);
			header('Content-Type: '.$mime_type);
			//3,写入正确的附件文件名头信息
			$orign_fname = $this->fileModel["orign_name"];
			$ua = $_SERVER["HTTP_USER_AGENT"]; // 处理不同浏览器的兼容性
			if (preg_match("/Firefox/", $ua)) {
				$encoded_filename = rawurlencode($orign_fname);
				header("Content-Disposition: attachment; filename*=\"utf8''" . $encoded_filename . '"');
			} else if (preg_match("/MSIE/", $ua) || preg_match("/Edge/", $ua) || preg_match("/rv:/", $ua)) {
				$encoded_filename = rawurlencode($orign_fname);
				header('Content-Disposition: attachment; filename="' . $encoded_filename . '"');
			} else {
				// for Chrome,Safari etc.
				header('Content-Disposition: attachment;filename="'. $orign_fname .'";filename*=utf-8'."''". $orign_fname);
			}
			exit;
		}else{
			$filePath = str_replace("\\","/",$filePath);
			header('Content-Type: '.self::getMimetype($realPath)); 
			if($header == "X-Accel-Redirect"){
				ob_flush();
				flush();
				echo "s";
			}
			header($header.": ".str_replace('%2F', '/', rawurlencode($filePath)));
			ob_flush();
			flush();
		}
    }
    
    /**
     * 无限速发送文件数据
     *
     * @param boolean $download 是否为下载
     * @param boolean $reload   是否支持断点续传
     * @return void
     */
    public function outputWithoutLimit($download = false,$reload = false){
		ignore_user_abort(false);
		$filePath = ROOT_PATH . 'public/uploads/' . $this->fileModel["pre_name"];
		set_time_limit(0);
		session_write_close();
		$file_size = filesize($filePath);  
		$ranges = $this->getRange($file_size);  
		if($reload == 1 && $ranges!=null){
			header('HTTP/1.1 206 Partial Content');  
			header('Accept-Ranges:bytes');  
			header(sprintf('content-length:%u',$ranges['end']-$ranges['start']));  
			header(sprintf('content-range:bytes %s-%s/%s', $ranges['start'], $ranges['end']-1, $file_size));  
		} 
		if($download){
			header('Cache-control: private');
			header('Content-Type: application/octet-stream'); 
			header('Content-Length: '.filesize($filePath)); 
			$encoded_fname = rawurlencode($this->fileModel["orign_name"]);
			header('Content-Disposition: attachment;filename="'.$encoded_fname.'";filename*=utf-8'."''".$encoded_fname); 
			ob_flush();
			flush();
		}
		if(file_exists($filePath)){
			if(!$download){
				header('Content-Type: '.self::getMimetype($filePath)); 
				ob_flush();
				flush();
			}
			$fileObj = fopen($filePath,"rb");
			if($reload == 1){
				fseek($fileObj, sprintf('%u', $ranges['start']));
			}
			while(!feof($fileObj)){
				echo fread($fileObj,10240);
				ob_flush();
				flush();
			} 
			fclose($fileObj);
		}
	}

    /**
     * 有限速发送文件数据
     *
     * @param int $speed        最大速度
     * @param boolean $download 是否为下载请求
     * @return void
     */
	public function outputWithLimit($speed,$download = false){
		ignore_user_abort(false);
		$filePath = ROOT_PATH . 'public/uploads/' . $this->fileModel["pre_name"];
		set_time_limit(0);
		session_write_close();
		if($download){
			header('Cache-control: private');
			header('Content-Type: application/octet-stream'); 
			header('Content-Length: '.filesize($filePath)); 
			$encoded_fname = rawurlencode($this->fileModel["orign_name"]);
			header('Content-Disposition: attachment;filename="'.$encoded_fname.'";filename*=utf-8'."''".$encoded_fname); 
			ob_flush();
			flush();
		}else{
			header('Content-Type: '.self::getMimetype($filePath)); 
			ob_flush();
			flush();
		}
		if(file_exists($filePath)){
			$fileObj = fopen($filePath,"r");
			while (!feof($fileObj)){ 
				echo fread($fileObj,round($speed*1024));
				ob_flush();
				flush();
				sleep(1);
			} 
			fclose($fileObj);
		}
	}

    /**
     * 获取文件MIME Type
     *
     * @param string $path 文件路径
     * @return void
     */
	static function getMimetype($path){
		//FILEINFO_MIME will output something like "image/jpeg; charset=binary"
		$finfoObj	= finfo_open(FILEINFO_MIME);
		$mimetype = finfo_file($finfoObj, $path);
		finfo_close($finfoObj);
		return $mimetype;
    }
    
    /**
     * 获取文件MIME Type
     *
     * @param string $path 文件路径
     * @return void
     */
	static function getMimetypeOnly($path){
		//FILEINFO_MIME_TYPE will output something like "image/jpeg"
		$finfoObj	= finfo_open(FILEINFO_MIME_TYPE);
		$mimetype = finfo_file($finfoObj, $path);
		finfo_close($finfoObj);
		return $mimetype;
	}

    /**
     * 获取断点续传时HTTP_RANGE头
     *
     * @param int $file_size 文件大小
     * @return void
     */
	private function getRange($file_size){  
		if(isset($_SERVER['HTTP_RANGE']) && !empty($_SERVER['HTTP_RANGE'])){  
			$range = $_SERVER['HTTP_RANGE'];  
			$range = preg_replace('/[\s|,].*/', '', $range);  
			$range = explode('-', substr($range, 6));  
			if(count($range)<2){  
				$range[1] = $file_size;  
			}  
			$range = array_combine(array('start','end'), $range);  
			if(empty($range['start'])){  
				$range['start'] = 0;  
			}  
			if(empty($range['end'])){  
				$range['end'] = $file_size;  
			}  
			return $range;  
		}  
		return null;  
    }

    /**
     * 生成/返回 文件缩略图
     *
     * @return array 重定向信息
     */
    public function getThumb(){
		$picInfo = explode(",",$this->fileModel["pic_info"]);
		$picInfo = self::getThumbSize($picInfo[0],$picInfo[1]);
		if(file_exists(ROOT_PATH . "public/thumb/".$this->fileModel["pre_name"]."_thumb")){
			self::outputThumb(ROOT_PATH . "public/thumb/".$this->fileModel["pre_name"]."_thumb");
			return [0,0];
		}
		$thumbImg = new Thumb(ROOT_PATH . "public/uploads/".$this->fileModel["pre_name"]);
		$thumbImg->thumb($picInfo[1], $picInfo[0]);
		if(!is_dir(dirname(ROOT_PATH . "public/thumb/".$this->fileModel["pre_name"]))){
			mkdir(dirname(ROOT_PATH . "public/thumb/".$this->fileModel["pre_name"]),0777,true);
		}
		$thumbImg->out(ROOT_PATH . "public/thumb/".$this->fileModel["pre_name"]."_thumb");
		self::outputThumb(ROOT_PATH . "public/thumb/".$this->fileModel["pre_name"]."_thumb");
		return [0,0];
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
     * 输出缩略图
     *
     * @param string $path 缩略图文件路径
     * @return void
     */
    static function outputThumb($path){
		ob_end_clean();
		if(!input("get.cache")=="no"){
			header("Cache-Control: max-age=10800");
		}
		header('Content-Type: '.self::getMimetype($path)); 
		$fileObj = fopen($path,"r");
		echo fread($fileObj,filesize($path)); 
		fclose($file); 
    }
    
    /**
     * 处理下载请求
     *
     * @param boolean $isAdmin 是否为管理员请求
     * @return void
     */
    public function Download($isAdmin=false){
		$speedLimit = Db::name('groups')->where('id',$this->userModel["user_group"])->find();
		$rangeTransfer = $speedLimit["range_transfer"];
		$speedLimit = $speedLimit["speed"];
		$sendFileOptions = Option::getValues(["download"]);
		if($sendFileOptions["sendfile"] == "1"){
			$this->sendFile($speedLimit,$rangeTransfer,true,$sendFileOptions["header"]);
		}else{
			if($isAdmin){
				$speedLimit = "";
			}
			if($speedLimit == "0"){
				exit();
			}else if(empty($speedLimit)){
				$this->outputWithoutLimit(true,$rangeTransfer);
				exit();
			}else if((int)$speedLimit > 0){
				$this->outputWithLimit($speedLimit,true);
			}
		}
    }
    
    /**
	 * 删除指定本地文件
	 *
	 * @param array $fileList   待删除文件的数据库记录
	 * @param array $policyData 待删除文件的上传策略信息
	 * @return void
	 */
	static function DeleteFile($fileList,$policyData){
		$fileListTemp = array_column($fileList, 'pre_name'); 
		foreach ($fileListTemp as $key => $value) {
			@unlink(ROOT_PATH . 'public/uploads/'.$value);
			if(file_exists(ROOT_PATH . 'public/thumb/'.$value."_thumb")){
				@unlink(ROOT_PATH . 'public/thumb/'.$value."_thumb");
			}
		}
    }
    
    /**
     * 签名临时直链，用于Office365预览
     *
     * @return array
     */
    public function signTmpUrl(){
        $options = Option::getValues(["oss","basic"]);
		$timeOut = $options["timeout"];
		$delayTime = time()+$timeOut;
		$key=$this->fileModel["id"].":".$delayTime.":".md5($this->userModel["user_pass"].$this->fileModel["id"].$delayTime.config("salt"));
		return [1,$options['siteURL']."Callback/TmpPreview/key/".$key];
    }

}

?>