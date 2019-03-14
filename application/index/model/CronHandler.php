<?php
namespace app\index\model;

use think\Model;
use think\Db;
use \think\Session;
use \app\index\model\FileManage;
use \app\index\model\Option;
use \app\index\model\Mail;
use \app\index\model\Aria2;
use think\Exception;

class CronHandler extends Model{

	public $cornTasks;
	public $timeNow;
	public $notifiedUid = [];

	public function __construct(){
		$this->cornTasks = Db::name('corn')->where('enable',1)->order('rank desc')->select();
		$this->timeNow = time();
	}

	public function checkInterval($interval,$last){
		return ($last+$interval)<= $this->timeNow ? true : false;
	}

	public function setComplete($name){
		Db::name('corn')->where('name', $name)->update(['last_excute' => $this->timeNow]);
	}

	public function Doit(){
		foreach ($this->cornTasks as $key => $value) {
			switch ($value["name"]) {
				case 'delete_unseful_chunks':
					if($this->checkInterval($value["interval_s"],$value["last_excute"])){
						$this->deleteUnsefulChunks($value["interval_s"]);
					}
					break;
				case 'delete_callback_data':
					if($this->checkInterval($value["interval_s"],$value["last_excute"])){
						$this->deleteCallbackData($value["interval_s"]);
					}
					break;
				case 'flush_aria2':
					if($this->checkInterval($value["interval_s"],$value["last_excute"])){
						$this->flushAria2($value["interval_s"]);
					}
					break;
				case 'flush_onedrive_token':
					if($this->checkInterval($value["interval_s"],$value["last_excute"])){
						$this->flushOnedriveToken($value["interval_s"]);
					}
					break;
				case 'delete_remote_downloader_failed_folder':
					if($this->checkInterval($value["interval_s"],$value["last_excute"])){
						$this->deleteRemoteDownloaderFailedFolder($value["interval_s"]);
					}
					break;
				default:
					# code...
					break;
			}
		}
	}

	private function deleteUnsefulChunks($interval){
		echo("deleteUnsefulChunks...");
		$chunkInfo = Db::name('chunks')->whereTime('time', '<', date('Y-m-d', time()-86400))->select();
		$deleteList=[];
		foreach ($chunkInfo as $key => $value) {
			$fileSize = @filesize(ROOT_PATH . 'public/uploads/chunks/'.$value["obj_name"].".chunk");
			@unlink(ROOT_PATH . 'public/uploads/chunks/'.$value["obj_name"].".chunk");
			FileManage::storageGiveBack($value["user"],$fileSize?$fileSize:0);
			$deleteList["$key"] = $value["id"];
		}
		Db::name('chunks')->where(['id' => ["in",$deleteList],])->delete();
		$this->setComplete("delete_unseful_chunks");
		echo("Complete<br>"); 
	}

	private function deleteCallbackData($interval){
		echo("deleteCallbackData...");
		Db::name("callback")->delete(true);
		echo("Complete<br>");
		$this->setComplete("delete_callback_data");
	}

	public function flushAria2($interval){
		echo("flushingAria2Status...");
		$aria2Options = Option::getValues(["aria2"]);
		$aria2 = new Aria2($aria2Options);
		$toBeFlushed = Db::name("download")
		->where("status","<>","complete")
		->where("status","<>","error")
		->where("status","<>","canceled")
		->select();
		foreach ($toBeFlushed as $key => $value) {
			$aria2->flushStatus($value["id"],$value["owner"],null);
		}
		echo("Complete<br>");
		$this->setComplete("flush_aria2");
	}

	public function deleteRemoteDownloaderFailedFolder($interval){
		echo("flushOnedriveToken...");
		$toBeDeleted = Db::name("task")
			->where("type","deleteFolder")
			->where("status","todo")
			->select();
		$success=[];
		$todo = [];
		foreach ($toBeDeleted as $key => $value) {
			$attr = json_decode($value["attr"],true);
			if(file_exists($attr["folder"])){
				self::remove_directory($attr["folder"]);
			}
			if(file_exists($attr["folder"])){
				$todo[] = $value["id"];
			}else{
				$success[] = $value["id"];
			}
		}

		Db::name("task")->where("id","in",$success)->update(["status"=>"success"]);

	}

	/**
	 * 移除整个目录
	 *
	 * @param string $dir
	 * @return void
	 */
	static function remove_directory($dir){
		if($handle=opendir("$dir")){
			while(false!==($item=readdir($handle))){
				if($item!="."&&$item!=".."){
					if(is_dir("$dir/$item")){
						self::remove_directory("$dir/$item");
					}else{
						unlink("$dir/$item");
					}
				}
			}
			closedir($handle);
			rmdir($dir);
		}
	}

	public function flushOnedriveToken($interval){
		echo("flushOnedriveToken...");
		$toBeFlushedPolicy = Db::name("policy")->where("policy_type","onedrive")->select();
		foreach ($toBeFlushedPolicy as $key => $value) {
			$onedrive = new \Krizalys\Onedrive\Client([
				'stream_back_end' => \Krizalys\Onedrive\StreamBackEnd::TEMP,
				'client_id' => $value["bucketname"],
			
				// Restore the previous state while instantiating this client to proceed in
				// obtaining an access token.
				'state' => json_decode($value["sk"]),
			]);
			try{
				$onedrive->renewAccessToken($value["ak"]);
			}catch(\Exception $e){

			}
			Db::name("policy")->where("id",$value["id"])->update([
				"sk" => json_encode($onedrive->getState()),
			]);
		}
		echo("Complete<br>");
		$this->setComplete("flush_onedrive_token");
	}

}
?>