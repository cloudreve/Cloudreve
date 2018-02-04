<?php
namespace app\index\model;

use think\Model;
use think\Db;
use \think\Session;
use \app\index\model\FileManage;
use \app\index\model\Option;
use \app\index\model\Mail;

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

}
?>