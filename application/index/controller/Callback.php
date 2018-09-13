<?php
namespace app\index\controller;

use think\Controller;
use think\Db;
use think\Request;
use \app\index\model\CallbackHandler;
use \app\index\model\User;
use \app\index\model\Option;
use \app\index\model\FileManage;
use think\Session;

class Callback extends Controller{

	public function index(){
		return "";
	}
	
	public function Qiniu(){
		ob_end_clean();
		header('Content-Type: application/json');
		$handllerObj = new CallbackHandler(file_get_contents("php://input"));
		$handllerObj -> qiniuHandler(Request::instance()->header('Authorization'));
	}

	public function Oss(){
		ob_end_clean();
		error_log("sadasdasdsadsasadasasdasdasd");
		header('Content-Type: application/json');
		$handllerObj = new CallbackHandler(file_get_contents("php://input"));
		$handllerObj -> ossHandler(Request::instance()->header('Authorization'),Request::instance()->header('x-oss-pub-key-url'));
	}

	public function TmpPreview(){
		$basicOptions = Option::getValues(['basic']);
		$params = explode(":",input("param.key"));
		$fileData = Db::name("files")->where("id",$params[0])->find();
		if (empty($fileData)){
			abort(404);
		}
		$userData = Db::name("users")->where("id",$fileData["upload_user"])->find();
		if(md5($userData["user_pass"].$fileData["id"].$params[1].config("salt")) != $params[2] || time()<$params[0]){
			abort(403);
		}
		$fileObj = new FileManage(rtrim($fileData["dir"],"/")."/".$fileData["orign_name"],$userData["id"]);
		$Redirect = $fileObj->PreviewHandler();
	}

	public function Upyun(){
		$signToken = Request::instance()->header('Authorization');
		$reqDate = Request::instance()->header('Date');
		$contentMd5 = Request::instance()->header('Content-MD5');
		ob_end_clean();
		header('Content-Type: application/json');
		$callbackData = [
			"code" => input("post.code"),
			"file_size" => input("post.file_size"),
			"url" => input("post.url"),
			"image-width" => input("post.image-width"),
			"image-height" => input("post.image-height"),
			"ext-param" => json_decode(input("post.ext-param"),true),
		];
		$handllerObj = new CallbackHandler($callbackData);
		$handllerObj -> upyunHandler($signToken,$reqDate,$contentMd5);
	}

	public function S3(){
		$request = Request::instance();
		if($request->method() == "OPTIONS"){
			ob_end_clean();
			header("Access-Control-Allow-Origin: *");
			exit();
		}
		ob_end_clean();
		header("Access-Control-Allow-Origin: *");
		$callbackKey = input("param.key");
		$callbackData = [
			"bucket" => input("get.bucket"),
			"key" => input("get.key"),
		];
		$handllerObj = new CallbackHandler($callbackData);
		$handllerObj -> s3Handler($callbackKey);
	}

	public function Remote(){
		ob_end_clean();
		header('Content-Type: application/json');
		$handllerObj = new CallbackHandler(file_get_contents("php://input"));
		$handllerObj -> remoteHandler(Request::instance()->header('Authorization'));
	}

}
