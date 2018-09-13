<?php
if(file_exists('../application/database.php')){
  echo "application/database.php 已存在，请备份并删除后再试";
  exit();
}
if(isset($_POST["mysqlServer"])){
  error_reporting(0);
  header('Content-Type:application/json; charset=utf-8');
  if(!file_exists('../mysql.sql')){
    echo json_encode(["error"=>true,"msg"=>"找不到mysql.sql"]);
    exit();
  }
  $sqlSource = file_get_contents('../mysql.sql');
  $sqlSource = str_replace("https://cloudreve.org/", $_POST["siteUrl"], $sqlSource);
  $mysqli = @new \mysqli($_POST["mysqlServer"], $_POST["mysqlUser"],  $_POST["mysqlPwd"], $_POST["mysqlDb"],  (int)$_POST["mysqlPort"]);
		if ($mysqli->connect_error) {
			@$mysqli->close();
			echo json_encode(["error"=>true,"msg"=>$mysqli->connect_error]);
      exit();
    }
    if (!$mysqli->multi_query($sqlSource)) {
      echo json_encode(["error"=>true,"msg"=>"无法写入数据表"]);
      exit();
    }
    if(file_exists('../application/database.php')){
      echo json_encode(["error"=>true,"msg"=>"application/database.php 已存在，请备份并删除后再试"]);
      exit();
    }
    try {
      $fileContent = file_get_contents("database_sample.php");
      $replacement = array(
       '{hostname}' => $_POST["mysqlServer"],
       '{database}' => $_POST["mysqlDb"],
       '{username}' => $_POST["mysqlUser"],
       '{password}' => $_POST["mysqlPwd"],
       '{hostport}' => $_POST["mysqlPort"],
       );
     $fileContent = strtr($fileContent,$replacement);
     file_put_contents('../application/database.php',$fileContent);
   }catch (Exception $e) {
    echo json_encode(["error"=>true,"msg"=>"无法写入数据库配置文件"]);
    exit();
   }
   echo json_encode(["error"=>false,"msg"=>""]);
  exit();
}

$phpVersionCheck = version_compare(PHP_VERSION,'5.6.0', '>');
$success = '<span style="color: #009688;"><i class="fa fa-check-circle" aria-hidden="true"></i> 满足</span>';
$error = '<span style="color: #F44336;"><i class="fa fa-times-circle" aria-hidden="true"></i> 不满足</span>';

$runtimeDirCheck = is_writable("../runtime");
$applicationDirCheck = is_writable("../application");
$publicDownloadsDirCheck = is_writable("../public/downloads");
$publicAvatarsDirCheck = is_writable("../public/avatars");
$publicThumbDirCheck = is_writable("../public/thumb");
$publicUploadsDirCheck = is_writable("../public/uploads");
$publicUploadsChunksDirCheck = is_writable("../public/uploads/chunks");

$curlCheck = extension_loaded("curl");
$pdoCheck = extension_loaded("pdo");
$fileinfoCheck = extension_loaded("fileinfo");
$gdCheck = extension_loaded("gd");

$thinkCaptchaCheck = is_dir("../vendor/topthink/think-captcha");
$ossCheck = is_dir("../vendor/aliyuncs/oss-sdk-php");
$davCheck = is_dir("../vendor/sabre/dav");
$upyunCheck = is_dir("../vendor/upyun/sdk");
$googleauthenticatorCheck = is_dir("../vendor/phpgangsta/googleauthenticator");
$qrcodeCheck = is_dir("../vendor/endroid/qrcode");

$isOk = $phpVersionCheck && $runtimeDirCheck && $applicationDirCheck && $publicAvatarsDirCheck && $curlCheck && $pdoCheck && $fileinfoCheck;
?>
<html lang="zh-cn" data-ng-app="FileManagerApp">
	<head>
		<meta name="viewport" content="initial-scale=1.0, user-scalable=no">
		<meta charset="utf-8">
		<meta name="theme-color" content="#4e64d9"/>
		<title>安装向导- Cloudreve</title>
		<!-- third party -->
		<script src="/static/js/jquery.min.js"></script>
		<link rel="stylesheet" href="/static/css/bootstrap.min.css" />
		<link rel="stylesheet" href="/static/css/material.css" />
		<script src="/static/js/material.js"></script>
		<script src="/static/js/bootstrap.min.js"></script>
		<link rel="stylesheet" href="/static/css/font-awesome.min.css">
		<!-- /third party -->
		<!-- Comment if you need to use raw source code -->
		<link href="/static/css/toastr.min.css" rel="stylesheet">
		<script type="text/javascript" src="/static/js/toastr.min.js"></script>
		<!-- /Comment if you need to use raw source code -->

    </head>
    <body data-ma-header="teal">
    <nav class="navbar navbar-inverse" style="background-color: rgb(78, 100, 217);">
            <div class="container-fluid">
                <div class="container">
<div class="navbar-header">
                        <div>
            <a class="navbar-brand waves-light waves-effect waves-block" href="/">
                
            </a>
</div>
                        
                        <button type="button" class="navbar-toggle collapsed" data-toggle="collapse" data-target="#bs-example-navbar-collapse-1" aria-expanded="false">
                        <span class="sr-only">Toggle navigation</span>
                        <span class="icon-bar"></span>
                        <span class="icon-bar"></span>
                        <span class="icon-bar"></span>
                        </button>
                        
                    </div>
                    <!-- Collect the nav links, forms, and other content for toggling -->
                    <div class="collapse navbar-collapse" id="bs-example-navbar-collapse-1">

              
                        </div><!-- /.navbar-collapse -->
                        </div><!-- /.container-fluid -->
                    </div>
                </nav>
<div class="container" id="enviromentCheck">
    <h1>环境检查</h1><br>
    <div class="panel panel-default">
  <div class="panel-heading">基本环境</div>
  <div class="panel-body">
  <table class="table table-hover ">
  <thead>
  <tr>
    <th>#</th>
    <th>项目</th>
    <th width="50%">说明</th>
    <th>必要性</th>
    <th>当前</th>
    <th>状态</th>
  </tr>
  </thead>
  <tbody>
  <tr <?php echo $phpVersionCheck?"":"class='danger'"; ?>>
    <td >1</td>
    <td>PHP版本 >= 5.6</td>
    <td>满足Cloudreve基本需求的最低PHP版本为5.6</td>
    <td>必须</td>
    <td><?php echo phpversion(); ?></td>
    <td><?php echo $phpVersionCheck?$success:$error; ?></td>
  </tr>
  <tr id="rewriteCheck">
    <td >2</td>
    <td>URL Rewrite</td>
    <td>服务器需正确配置URL重写规则（伪静态），否则各个页面将会返回404错误</td>
    <td>必须</td>
    <td id="rewriteStatus"></td>
    <td>
    <span id="rewriteSuccess" style="display:none"><?php echo $success?></span>
    <span id="rewriteError" style="display:none"><?php echo $error?></span>
    </td>
  </tr>
  </tbody>
</table>
  </div>
</div>




<div class="panel panel-default">
  <div class="panel-heading">读写权限</div>
  <div class="panel-body">
  <table class="table table-hover ">
  <thead>
  <tr>
    <th>#</th>
    <th>目录</th>
    <th width="50%">说明</th>
    <th>必要性</th>
    <th>状态</th>
  </tr>
  </thead>
  <tbody>
  <tr <?php echo $runtimeDirCheck?"":"class='danger'"; ?>>
    <td >1</td>
    <td>runtime 可读写</td>
    <td>runtime用于存放系统工作产生的临时文件、日志、缓存等数据</td>
    <td>必须</td>
    <td><?php echo $runtimeDirCheck?$success:$error; ?></td>
  </tr>
  <tr <?php echo $applicationDirCheck?"":"class='danger'"; ?>>
    <td >2</td>
    <td>application 可读写</td>
    <td>application用于安装程序写入数据库配置文件，仅安装时需要写入权限</td>
    <td>必须(临时)</td>
    <td><?php echo $applicationDirCheck?$success:$error; ?></td>
  </tr>
  <tr <?php echo $publicAvatarsDirCheck?"":"class='danger'"; ?>>
    <td >3</td>
    <td>public/avatars 可读写</td>
    <td>用于存放用户头像</td>
    <td>必须</td>
    <td><?php echo $publicAvatarsDirCheck?$success:$error; ?></td>
  </tr>
  <tr <?php echo $publicUploadsDirCheck?"":"class='warning'"; ?>>
    <td >4</td>
    <td>public/uploads 可读写</td>
    <td>用于存放本地策略上传的文件数据</td>
    <td>可选</td>
    <td><?php echo $publicUploadsDirCheck?$success:$error; ?></td>
  </tr>
  <tr <?php echo $publicUploadsChunksDirCheck?"":"class='warning'"; ?>>
    <td >5</td>
    <td>public/uploads/chunks 可读写</td>
    <td>用于存放本地策略上传文件的临时分片数据</td>
    <td>可选</td>
    <td><?php echo $publicUploadsChunksDirCheck?$success:$error; ?></td>
  </tr>
  <tr <?php echo $publicDownloadsDirCheck?"":"class='warning'"; ?>>
    <td >6</td>
    <td>public/downloads 可读写</td>
    <td>用于存放离线下载的文件数据</td>
    <td>可选</td>
    <td><?php echo $publicDownloadsDirCheck?$success:$error; ?></td>
  </tr>
  </tbody>
</table>
  </div>
</div>






<div class="panel panel-default">
  <div class="panel-heading">PHP扩展</div>
  <div class="panel-body">
  <table class="table table-hover ">
  <thead>
  <tr>
    <th>#</th>
    <th>扩展名</th>
    <th width="50%">说明</th>
    <th>必要性</th>
    <th>状态</th>
  </tr>
  </thead>
  <tbody>
  <tr <?php echo $curlCheck?"":"class='danger'"; ?>>
    <td >1</td>
    <td>curl</td>
    <td>发送网络请求</td>
    <td>必须</td>
    <td><?php echo $curlCheck?$success:$error; ?></td>
  </tr>

  <tr <?php echo $pdoCheck?"":"class='danger'"; ?>>
    <td >2</td>
    <td>pdo</td>
    <td>数据库操作</td>
    <td>必须</td>
    <td><?php echo $pdoCheck?$success:$error; ?></td>
  </tr>
  <tr <?php echo $fileinfoCheck?"":"class='warnging'"; ?>>
    <td >3</td>
    <td>fileinfo</td>
    <td>用于处理本地策略图像文件预览、用户头像展示</td>
    <td>推荐</td>
    <td><?php echo $fileinfoCheck?$success:$error; ?></td>
  </tr>
  <tr <?php echo $gdCheck?"":"class='warnging'"; ?>>
    <td >4</td>
    <td>gd</td>
    <td>用于生成验证码</td>
    <td>推荐</td>
    <td><?php echo $gdCheck?$success:$error; ?></td>
  </tr>
  </tbody>
</table>
  </div>
</div>



<div class="panel panel-default">
  <div class="panel-heading">依赖库</div>
  <div class="panel-body">
  <table class="table table-hover ">
  <thead>
  <tr>
    <th>#</th>
    <th>库名</th>
    <th width="50%">说明</th>
    <th>必要性</th>
    <th>状态</th>
  </tr>
  </thead>
  <tbody>
  <tr <?php echo $thinkCaptchaCheck?"":"class=''"; ?>>
    <td >1</td>
    <td>think-captcha</td>
    <td>生成验证码图像</td>
    <td>可选</td>
    <td><?php echo $thinkCaptchaCheck?$success:$error; ?></td>
  </tr>

  <tr <?php echo $ossCheck?"":"class=''"; ?>>
    <td >2</td>
    <td>oss-sdk-php</td>
    <td>阿里云OSS上传策略需要使用</td>
    <td>可选</td>
    <td><?php echo $ossCheck?$success:$error; ?></td>
  </tr>
  <tr <?php echo $davCheck?"":"class=''"; ?>>
    <td >3</td>
    <td>dav</td>
    <td>WebDAV功能需要使用</td>
    <td>可选</td>
    <td><?php echo $davCheck?$success:$error; ?></td>
  </tr>
  <tr <?php echo $upyunCheck?"":"class=''"; ?>>
    <td >4</td>
    <td>upyun/sdk</td>
    <td>又拍云上传策略需要使用</td>
    <td>可选</td>
    <td><?php echo $upyunCheck?$success:$error; ?></td>
  </tr>
  <tr <?php echo $googleauthenticatorCheck?"":"class=''"; ?>>
    <td >5</td>
    <td>googleauthenticator</td>
    <td>二步验证</td>
    <td>可选</td>
    <td><?php echo $googleauthenticatorCheck?$success:$error; ?></td>
  </tr>
  <tr <?php echo $qrcodeCheck?"":"class=''"; ?>>
    <td >5</td>
    <td>endroid/qrcode</td>
    <td>用于生成二步验证的二维码</td>
    <td>可选</td>
    <td><?php echo $qrcodeCheck?$success:$error; ?></td>
  </tr>
  </tbody>
</table>

  </div>
</div>



<div style="text-align:right;"><button class="btn btn-lg btn-primary btn-raised" id="doInstall"><?php echo $isOk?"下一步":"忽略问题，继续下一步"; ?></button></div>
       </div>
       <div class="container" id="installSuccess" style="display:none">
       <div class="jumbotron">
       <h2>安装完成</h2>
       <p>您的Cloudreve站点初始管理员信息如下，请登陆后修改默认密码和邮箱。</p>
       <div class="form-group">
            <label for="adminUrl" class="col-md-2 control-label">管理后台地址</label>

            <div class="col-md-10">
                <input type="text" class="form-control" id="adminUrl" name="adminUrl" value="<?php  $url='http://'.$_SERVER['SERVER_NAME'].$_SERVER["REQUEST_URI"]; 
$mulu= dirname($url);
echo $mulu."/Admin";
?>">
            </div>
        </div>


 <div class="form-group">
            <label for="admin" class="col-md-2 control-label">管理员账号</label>

            <div class="col-md-10">
                <input type="text" class="form-control" id="admin" name="admin" value="admin@cloudreve.org">
            </div>
        </div>
        <div class="form-group">
            <label for="adminPwd" class="col-md-2 control-label">管理员密码</label>

            <div class="col-md-10">
                <input type="text" class="form-control" id="adminPwd" name="adminPwd" value="admin">
            </div>
        </div>
        <br><br><br><br>
       </div>
       </div>

<div class="container" id="installForm" style="display:none">
<div class="jumbotron">
<h2>信息填写</h2>
    <form id="setUpInfo">
        <div class="form-group">
            <label for="siteUrl" class="col-md-2 control-label">站点URL</label>

            <div class="col-md-10">
                <input type="text" class="form-control" id="siteUrl" name="siteUrl" placeholder="结尾需要加 / " value="<?php  $url='http://'.$_SERVER['SERVER_NAME'].$_SERVER["REQUEST_URI"]; 
$mulu= dirname($url);
echo $mulu."/";
?>">
            </div>
        </div>

        <div class="form-group">
            <label for="mysqlServer" class="col-md-2 control-label">MySQL服务器</label>

            <div class="col-md-10">
                <input type="text" class="form-control" id="mysqlServer" name="mysqlServer" value="localhost">
            </div>
        </div>
        <div class="form-group">
            <label for="mysqlPort" class="col-md-2 control-label">MySQL端口</label>

            <div class="col-md-10">
                <input type="number" class="form-control" id="mysqlPort" name="mysqlPort" value="3306">
            </div>
        </div>
        <div class="form-group">
            <label for="mysqlUser" class="col-md-2 control-label">MySQL用户名</label>

            <div class="col-md-10">
                <input type="text" class="form-control" id="mysqlUser" name="mysqlUser" value="root">
            </div>
        </div>
        <div class="form-group">
            <label for="mysqlPwd" class="col-md-2 control-label">MySQL密码</label>

            <div class="col-md-10">
                <input type="password" class="form-control" id="mysqlPwd" name="mysqlPwd">
            </div>
        </div>
        <div class="form-group">
            <label for="mysqlDb" class="col-md-2 control-label">数据库名</label>

            <div class="col-md-10">
                <input type="text" class="form-control" id="mysqlDb" name="mysqlDb" >
            </div>
        </div><br>
        <div style="text-align:right;"><button type="button" class="btn btn-lg btn-primary btn-raised" id="startInstall">开始安装</button></div>
       </div>
<br><br><br><br>
    </form>
</div>
</div>

            </body>

            <script type="text/javascript">
        

            </script>

<script type="text/javascript">
$.material.init();
$.get("/Member", function(result){
    $("#rewriteStatus").html("正常");
    $("#rewriteSuccess").show();
}).error(function(){
    $("#rewriteStatus").html("异常");
    $("#rewriteError").show();
    $("#rewriteCheck").addClass("danger");
});
$("#doInstall").click(function(){
    $("#enviromentCheck").fadeOut();
    $("#installForm").fadeIn();
})
$("#startInstall").click(function(){
  $.post("index.php",$("#setUpInfo").serialize(),function(data){
    console.log(data);
    if(data.error == true){
      toastr["error"](data.msg);
    }else{
      $("#installForm").fadeOut();
      $("#installSuccess").fadeIn();
    }
  }).error(function(){
    toastr["error"]("安装出现未知错误");
  })
})
</script>            
