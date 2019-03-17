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
		<script src="/static/js/jquery.min.js"></script>
		<link rel="stylesheet" href="/static/css/font-awesome.min.css">
		<link href="/static/css/toastr.min.css" rel="stylesheet">
		<script type="text/javascript" src="/static/js/toastr.min.js"></script>
<style>
html{font-size:10px;-webkit-tap-highlight-color:rgba(0,0,0,0)}.h1,.h2,.h3,.h4,body,h1,h2,h3,h4,h5,h6{font-weight:300}body{font-family:"Helvetica Neue",Helvetica,Arial,sans-serif;font-size:14px;line-height:1.42857143;color:#333;background-color:#fff}body{background-color:#EEE}.navbar{border:0;border-radius:0}.navbar{position:relative;min-height:50px;margin-bottom:20px}article,aside,details,figcaption,figure,footer,header,hgroup,main,menu,nav,section,summary{display:block}.container-fluid{padding-right:15px;padding-left:15px;margin-right:auto;margin-left:auto}@media(min-width:768px){.container{width:750px}}.container{padding-right:15px;padding-left:15px;margin-right:auto;margin-left:auto}@media(min-width:768px){.container-fluid>.navbar-collapse,.container-fluid>.navbar-header,.container>.navbar-collapse,.container>.navbar-header{margin-right:0;margin-left:0}}@media(min-width:768px){.navbar-header{float:left}}@media(min-width:768px){.navbar>.container .navbar-brand,.navbar>.container-fluid .navbar-brand{margin-left:-15px}}@media(max-width:1199px){.navbar .navbar-brand{height:50px;padding:24px 100px 35px}}.navbar .navbar-brand{position:relative;line-height:30px;color:inherit;    padding: 24px 100px 35px;}.navbar-brand{background-position:10px;background-size:192px 50px;background-image:url(/static/img/logo_s.png);width:200px;background-repeat:no-repeat}.navbar .navbar-collapse,.navbar .navbar-form{border-color:rgba(0,0,0,.1)}@media(min-width:768px){.container-fluid>.navbar-collapse,.container-fluid>.navbar-header,.container>.navbar-collapse,.container>.navbar-header{margin-right:0;margin-left:0}}@media(min-width:768px){.navbar-collapse.collapse{display:block!important;height:auto!important;padding-bottom:0;overflow:visible!important}}.navbar-collapse{padding-right:15px;padding-left:15px}.h1,h1{font-size:36px}.h1,.h2,.h3,h1,h2,h3{margin-top:20px;margin-bottom:10px}.h1,.h2,.h3,.h4,.h5,.h6,h1,h2,h3,h4,h5,h6{font-family:inherit;font-weight:500;line-height:1.1;color:inherit}h1{margin:.67em 0;font-size:2em}.h1,.h2,.h3,.h4,body,h1,h2,h3,h4,h5,h6{font-weight:300}@media(min-width:768px){.container{width:750px}}.container{padding-right:15px;padding-left:15px;margin-right:auto;margin-left:auto}body{margin:0}.panel-default{border-color:#ddd}.panel{margin-bottom:20px;background-color:#fff;border:1px solid transparent;border-radius:4px;-webkit-box-shadow:0 1px 1px rgba(0,0,0,.05);box-shadow:0 1px 1px rgba(0,0,0,.05)}.panel.panel-default>.panel-heading,.panel>.panel-heading{background-color:#eee}.panel-default>.panel-heading,.panel:not([class*=panel-])>.panel-heading{color:rgba(0,0,0,.87)}[class*=panel-]>.panel-heading{color:rgba(255,255,255,.84);border:0}.panel-default>.panel-heading{color:#333;background-color:#f5f5f5;border-color:#ddd}.panel-heading{padding:10px 15px;border-bottom:1px solid transparent;border-top-left-radius:3px;border-top-right-radius:3px}.panel-body{padding:15px}.panel{border-radius:2px;border:0;-webkit-box-shadow:0 1px 6px 0 rgba(0,0,0,.12),0 1px 6px 0 rgba(0,0,0,.12);box-shadow:0 1px 6px 0 rgba(0,0,0,.12),0 1px 6px 0 rgba(0,0,0,.12)}.panel-body{padding:15px}.table{width:100%;max-width:100%;margin-bottom:20px}table{background-color:transparent}table{border-spacing:0;border-collapse:collapse}.table>tbody>tr>td,.table>tbody>tr>th,.table>tfoot>tr>td,.table>tfoot>tr>th,.table>thead>tr>td,.table>thead>tr>th{padding:8px;line-height:1.42857143;vertical-align:top;border-top:1px solid #ddd}th{text-align:left}.table>caption+thead>tr:first-child>td,.table>caption+thead>tr:first-child>th,.table>colgroup+thead>tr:first-child>td,.table>colgroup+thead>tr:first-child>th,.table>thead:first-child>tr:first-child>td,.table>thead:first-child>tr:first-child>th{border-top:0}.table>thead>tr>th{vertical-align:bottom;border-bottom:2px solid #ddd}.table>tbody>tr>td,.table>tbody>tr>th,.table>tfoot>tr>td,.table>tfoot>tr>th,.table>thead>tr>td,.table>thead>tr>th{padding:8px;line-height:1.42857143;vertical-align:top;border-top:1px solid #ddd}.table-hover>tbody>tr:hover{background-color:#f5f5f5}.table>tbody>tr.danger>td,.table>tbody>tr.danger>th,.table>tbody>tr>td.danger,.table>tbody>tr>th.danger,.table>tfoot>tr.danger>td,.table>tfoot>tr.danger>th,.table>tfoot>tr>td.danger,.table>tfoot>tr>th.danger,.table>thead>tr.danger>td,.table>thead>tr.danger>th,.table>thead>tr>td.danger,.table>thead>tr>th.danger{background-color:#f2dede}.table>tbody>tr.warning>td,.table>tbody>tr.warning>th,.table>tbody>tr>td.warning,.table>tbody>tr>th.warning,.table>tfoot>tr.warning>td,.table>tfoot>tr.warning>th,.table>tfoot>tr>td.warning,.table>tfoot>tr>th.warning,.table>thead>tr.warning>td,.table>thead>tr.warning>th,.table>thead>tr>td.warning,.table>thead>tr>th.warning{background-color:#fcf8e3}button,input,select,textarea{font-family:inherit;font-size:inherit;line-height:inherit}.btn{display:inline-block;padding:6px 12px;margin-bottom:0;font-size:14px;font-weight:400;line-height:1.42857143;text-align:center;white-space:nowrap;vertical-align:middle;-ms-touch-action:manipulation;touch-action:manipulation;cursor:pointer;-webkit-user-select:none;-moz-user-select:none;-ms-user-select:none;user-select:none;background-image:none;border:1px solid transparent;border-radius:4px}
.btn-primary{color:#fff;background-color:#337ab7;border-color:#2e6da4}.btn-group-lg>.btn,.btn-lg{padding:10px 16px;font-size:18px;line-height:1.3333333;border-radius:6px}.btn,.input-group-btn .btn{border:0;border-radius:2px;position:relative;padding:8px 30px;margin:10px 1px;font-size:14px;font-weight:500;text-transform:uppercase;letter-spacing:0;will-change:box-shadow,transform;-webkit-transition:-webkit-box-shadow .2s cubic-bezier(.4,0,1,1),background-color .2s cubic-bezier(.4,0,.2,1),color .2s cubic-bezier(.4,0,.2,1);-o-transition:box-shadow .2s cubic-bezier(.4,0,1,1),background-color .2s cubic-bezier(.4,0,.2,1),color .2s cubic-bezier(.4,0,.2,1);transition:box-shadow .2s cubic-bezier(.4,0,1,1),background-color .2s cubic-bezier(.4,0,.2,1),color .2s cubic-bezier(.4,0,.2,1);outline:0;cursor:pointer;text-decoration:none;background:0}.btn-group-raised .btn.btn-primary,.btn-group-raised .input-group-btn .btn.btn-primary,.btn.btn-fab.btn-primary,.btn.btn-raised.btn-primary,.input-group-btn .btn.btn-fab.btn-primary,.input-group-btn .btn.btn-raised.btn-primary{background-color:#3f51b5;color:rgba(255,255,255,.84)}.btn-group-lg .btn,.btn-group-lg .input-group-btn .btn,.btn.btn-lg,.input-group-btn .btn.btn-lg{font-size:16px}.btn-group-raised .btn:not(.btn-link),.btn-group-raised .input-group-btn .btn:not(.btn-link),.btn.btn-raised:not(.btn-link),.input-group-btn .btn.btn-raised:not(.btn-link){-webkit-box-shadow:0 2px 2px 0 rgba(0,0,0,.14),0 3px 1px -2px rgba(0,0,0,.2),0 1px 5px 0 rgba(0,0,0,.12);box-shadow:0 2px 2px 0 rgba(0,0,0,.14),0 3px 1px -2px rgba(0,0,0,.2),0 1px 5px 0 rgba(0,0,0,.12)}.jumbotron{padding-top:30px;padding-bottom:30px;margin-bottom:30px;color:inherit;background-color:#eee}@media screen and (min-width:768px){.jumbotron{padding-top:48px;padding-bottom:48px}}.container .jumbotron,.container-fluid .jumbotron{padding-right:15px;padding-left:15px;border-radius:6px}@media screen and (min-width:768px){.container .jumbotron,.container-fluid .jumbotron{padding-right:60px;padding-left:60px}}body .container .jumbotron,body .container .well,body .container-fluid .jumbotron,body .container-fluid .well{background-color:#fff;padding:19px;margin-bottom:20px;-webkit-box-shadow:0 8px 17px 0 rgba(0,0,0,.2),0 6px 20px 0 rgba(0,0,0,.19);box-shadow:0 8px 17px 0 rgba(0,0,0,.2),0 6px 20px 0 rgba(0,0,0,.19);border-radius:2px;border:0}.form-group{padding-bottom:7px;margin:28px 0 0 0}.form-group{position:relative}label{display:inline-block;max-width:100%;margin-bottom:5px;font-weight:700}.checkbox label,.radio label,label{font-size:16px;line-height:1.42857143;color:#bdbdbd;font-weight:400}.col-lg-1,.col-lg-10,.col-lg-11,.col-lg-12,.col-lg-2,.col-lg-3,.col-lg-4,.col-lg-5,.col-lg-6,.col-lg-7,.col-lg-8,.col-lg-9,.col-md-1,.col-md-10,.col-md-11,.col-md-12,.col-md-2,.col-md-3,.col-md-4,.col-md-5,.col-md-6,.col-md-7,.col-md-8,.col-md-9,.col-sm-1,.col-sm-10,.col-sm-11,.col-sm-12,.col-sm-2,.col-sm-3,.col-sm-4,.col-sm-5,.col-sm-6,.col-sm-7,.col-sm-8,.col-sm-9,.col-xs-1,.col-xs-10,.col-xs-11,.col-xs-12,.col-xs-2,.col-xs-3,.col-xs-4,.col-xs-5,.col-xs-6,.col-xs-7,.col-xs-8,.col-xs-9{position:relative;min-height:1px;padding-right:15px;padding-left:15px}.form-group .checkbox label,.form-group .radio label,.form-group label{font-size:16px;line-height:1.42857143;color:#bdbdbd;font-weight:400}.form-group label.control-label{font-size:12px;line-height:1.07142857;font-weight:400;margin:16px 0 0 0}.col-lg-1,.col-lg-10,.col-lg-11,.col-lg-12,.col-lg-2,.col-lg-3,.col-lg-4,.col-lg-5,.col-lg-6,.col-lg-7,.col-lg-8,.col-lg-9,.col-md-1,.col-md-10,.col-md-11,.col-md-12,.col-md-2,.col-md-3,.col-md-4,.col-md-5,.col-md-6,.col-md-7,.col-md-8,.col-md-9,.col-sm-1,.col-sm-10,.col-sm-11,.col-sm-12,.col-sm-2,.col-sm-3,.col-sm-4,.col-sm-5,.col-sm-6,.col-sm-7,.col-sm-8,.col-sm-9,.col-xs-1,.col-xs-10,.col-xs-11,.col-xs-12,.col-xs-2,.col-xs-3,.col-xs-4,.col-xs-5,.col-xs-6,.col-xs-7,.col-xs-8,.col-xs-9{position:relative;min-height:1px;padding-right:15px;padding-left:15px}button,input,optgroup,select,textarea{margin:0;font:inherit;color:inherit}.form-control{display:block;width:100%;height:34px;padding:6px 12px;font-size:14px;line-height:1.42857143;color:#555;background-color:#fff;background-image:none;border:1px solid #ccc;border-radius:4px;-webkit-box-shadow:inset 0 1px 1px rgba(0,0,0,.075);box-shadow:inset 0 1px 1px rgba(0,0,0,.075);-webkit-transition:border-color ease-in-out .15s,-webkit-box-shadow ease-in-out .15s;-o-transition:border-color ease-in-out .15s,box-shadow ease-in-out .15s;transition:border-color ease-in-out .15s,box-shadow ease-in-out .15s}.form-control{height:38px;padding:7px 0;font-size:16px;line-height:1.42857143}.form-control,.form-group .form-control{border:0;background-image:-webkit-gradient(linear,left top,left bottom,from(#009688),to(#009688)),-webkit-gradient(linear,left top,left bottom,from(#d2d2d2),to(#d2d2d2));background-image:-webkit-linear-gradient(#009688,#009688),-webkit-linear-gradient(#d2d2d2,#d2d2d2);background-image:-o-linear-gradient(#009688,#009688),-o-linear-gradient(#d2d2d2,#d2d2d2);background-image:linear-gradient(#009688,#009688),linear-gradient(#d2d2d2,#d2d2d2);-webkit-background-size:0 2px,100% 1px;background-size:0 2px,100% 1px;background-repeat:no-repeat;background-position:center bottom,center -webkit-calc(100% - 1px);background-position:center bottom,center calc(100% - 1px);background-color:rgba(0,0,0,0);-webkit-transition:background 0s ease-out;-o-transition:background 0s ease-out;transition:background 0s ease-out;float:none;-webkit-box-shadow:none;box-shadow:none;border-radius:0}
.form-group .form-control{margin-bottom:7px}
</style>
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
