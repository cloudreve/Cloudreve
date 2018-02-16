$("#saveBasic").click(function() {
	$("#saveBasic").attr("disabled", "true");
	$.post("/Admin/SaveBasicSetting", 
		$("#basicOptionsForm").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#saveBasic").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("设置已保存");
			$("#saveBasic").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#saveBasic").removeAttr("disabled");
		}
	});
})

$("#qq_login1").click(function(){
	$("#qqLoginOptions").slideDown();
})
$("#qq_login2").click(function(){
	$("#qqLoginOptions").slideUp();
})
$("#saveReg").click(function() {
	$("#saveReg").attr("disabled", "true");
	$.post("/Admin/SaveRegSetting", 
		$("#basicOptionsForm").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#saveReg").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("设置已保存");
			$("#saveReg").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#saveReg").removeAttr("disabled");
		}
	});
})
$("#saveMail").click(function() {
	$("#saveMail").attr("disabled", "true");
	$.post("/Admin/SaveMailSetting", 
		$("#basicOptionsForm").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#saveMail").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("设置已保存");
			$("#saveMail").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#saveMail").removeAttr("disabled");
		}
	});
})

$("#sendMail").click(function() {
	$("#sendMail").attr("disabled", "true");
	$.post("/Admin/SendTestMail", 
		$("#testMail").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#sendMail").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("邮件发送成功");
			$("#sendMail").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#sendMail").removeAttr("disabled");
		}
	});
})

$("#saveTemplate").click(function() {
	$("#saveTemplate").attr("disabled", "true");
	$.post("/Admin/SaveMailTemplate", 
		$("#mailTemplate").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#saveTemplate").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("设置已保存");
			$("#saveTemplate").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#saveTemplate").removeAttr("disabled");
		}
	});
})

$("#saveJsj").click(function() {
	$("#saveJsj").attr("disabled", "true");
	$.post("/Admin/SaveMailTemplate", 
		$("#jsjFrom").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#saveJsj").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("设置已保存");
			$("#saveJsj").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#saveJsj").removeAttr("disabled");
		}
	});
})

$("#saveYz").click(function() {
	$("#saveYz").attr("disabled", "true");
	$.post("/Admin/SaveMailTemplate", 
		$("#yzFrom").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#saveYz").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("设置已保存");
			$("#saveYz").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#saveYz").removeAttr("disabled");
		}
	});
})
$("[name='sendfile']").change(function(){
  if($(this).val()=="1"){
  	$("#sendfile_header").slideDown();
  }else{
  	$("#sendfile_header").slideUp();
  }
});