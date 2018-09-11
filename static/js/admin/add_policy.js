$("#addLocal").click(function(){
	$("#choose").hide();
	$("#local").show();
})
$("#addQiniu").click(function(){
	$("#choose").hide();
	$("#qiniu").show();
})
$("#addOss").click(function(){
	$("#choose").hide();
	$("#oss").show();
})
$("#addUpyun").click(function(){
	$("#choose").hide();
	$("#upyun").show();
})
$("#local_allowd_origin1").click(function(){
	$("#localOrigin").slideDown();
})
$("#local_allowd_origin2").click(function(){
	$("#localOrigin").slideUp();
})
$("#autoname1").click(function(){
	$("#autoname_form").slideDown();
})
$("#autoname2").click(function(){
	$("#autoname_form").slideUp();
})
$("#qiniu_autoname1").click(function(){
	$("#qiniu_autoname_form").slideDown();
})
$("#qiniu_autoname2").click(function(){
	$("#qiniu_autoname_form").slideUp();
})
$("#oss_autoname1").click(function(){
	$("#oss_autoname_form").slideDown();
})
$("#oss_autoname2").click(function(){
	$("#oss_autoname_form").slideUp();
})
$("#localPolicy").submit(function() {
	$("#savePolicy").attr("disabled", "true");
	$.post("/Admin/SavePolicy", 
		$("#localPolicy").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#savePolicy").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("上传策略已添加");
			location.href = "/Admin/PolicyList?page=1";
		}else{
			toastr["warning"]("未知错误");
			$("#savePolicy").removeAttr("disabled");
		}
	});
	return false;
})
$("#onedrivePolicy").submit(function() {
	$("#savePolicy").attr("disabled", "true");
	$.post("/Admin/SavePolicy", 
		$("#onedrivePolicy").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#savePolicy").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("上传策略已添加");
			location.href = "/Admin/UpdateOnedriveToken?id="+data.id;
		}else{
			toastr["warning"]("未知错误");
			$("#savePolicy").removeAttr("disabled");
		}
	});
	return false;
})
$("#qiniuPolicy").submit(function() {
	$("#saveQiniu").attr("disabled", "true");
	$.post("/Admin/SavePolicy", 
		$("#qiniuPolicy").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#saveQiniu").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("上传策略已添加");
			location.href = "/Admin/PolicyList?page=1";
		}else{
			toastr["warning"]("未知错误");
			$("#saveQiniu").removeAttr("disabled");
		}
	});
	return false;
})
$("#bucket_private_1").click(function(){
	$("#qiniu_allowd_origin2").prop("checked","true");
	$("#outlink").slideUp();
})
$("#bucket_private_0").click(function(){
	$("#outlink").slideDown();
})
$("#upyun_bucket_private_1").click(function(){
	$("#upyun_allowd_origin2").prop("checked","true");
	$("#upyun_outlink").slideUp();
	$("#upyun_token").slideDown();
})
$("#upyun_bucket_private_0").click(function(){
	$("#upyun_outlink").slideDown();
	$("#upyun_token").slideUp();
})
$("#oss_private_1").click(function(){
	$("#oss_allowd_origin2").prop("checked","true");
	$("#oss_outlink").slideUp();
})
$("#oss_private_0").click(function(){
	$("#oss_outlink").slideDown();
})
$("#ossPolicy").submit(function() {
	$("#saveOss").attr("disabled", "true");
	$.post("/Admin/SavePolicy", 
		$("#ossPolicy").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#ossPolicy").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("上传策略已添加");
			location.href = "/Admin/PolicyList?page=1";
		}else{
			toastr["warning"]("未知错误");
			$("#ossPolicy").removeAttr("disabled");
		}
	});
	return false;
})
$("#upyunPolicy").submit(function() {
	$("#saveUpyun").attr("disabled", "true");
	$.post("/Admin/SavePolicy", 
		$("#upyunPolicy").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#saveUpyun").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("上传策略已添加");
			location.href = "/Admin/PolicyList?page=1";
		}else{
			toastr["warning"]("未知错误");
			$("#saveUpyun").removeAttr("disabled");
		}
	});
	return false;
})