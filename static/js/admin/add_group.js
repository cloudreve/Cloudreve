$("#addGroup").submit(function() {
	$("#saveGroup").attr("disabled", "true");
	$.post("/Admin/AddGroup", 
		$("#addGroup").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#saveGroup").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("用户组已添加");
			document.getElementById("addGroup").reset();
			$("#saveGroup").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#saveGroup").removeAttr("disabled");
		}
	});
	return false;
})
$("#editGroup").submit(function() {
	$("#saveGroup").attr("disabled", "true");
	$.post("/Admin/SaveGroup", 
		$("#editGroup").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#saveGroup").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("用户组已添加");
			location.href="/Admin/GroupList"
			$("#saveGroup").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#saveGroup").removeAttr("disabled");
		}
	});
	return false;
})