$("#dataTable_length").change(function() {
	var p1 = $(this).children('option:selected').val();
	$.cookie('pageSize', p1);
	location.href = "/Admin/Users?page=1";
});
$(document).ready(function() {
	var pageSize = ($.cookie('pageSize') == null) ? "10" : $.cookie('pageSize');
	var group = ($.cookie('userGroup') == null) ? "" : $.cookie('userGroup');
	var  searchColUser = ($.cookie('searchColUser') == null) ? "id" : $.cookie('searchColUser');
	$("#dataTable_length").val(pageSize);
	$("#searchFrom").val($.cookie('userSearch'));
	$("a[data-method='" + $.cookie('orderMethodUser') + "']").addClass("active");
	$("a[data-group='" + group + "']").addClass("active");
	$("a[data-status='" + $.cookie('userStatus') + "']").addClass("active");
	$("#searchColUser").val(searchColUser);
	$("#searchValueUser").val($.cookie('searchValueUser'));
})
$("#searchFrom").keydown(function(e) {
	var curKey = e.which;
	if (curKey == 13) {
		$.cookie('userSearch', $(this).val());
		location.href = "/Admin/Users?page=1";
	}
});
$("#applySearch").click(function(){
	$.cookie('searchColUser', $("#searchColUser").val());
	$.cookie('searchValueUser', $("#searchValueUser").val());
	location.href = "/Admin/Users?page=1";
})
$("#order").children().click(function() {
	$.cookie('orderMethodUser', $(this).children().attr("data-method"));
	location.href = "/Admin/Users?page=1";
})
$("#groupS").children().click(function() {
	$.cookie('userGroup', $(this).children().attr("data-group"));
	location.href = "/Admin/Users?page=1";
})
$("#status").children().click(function() {
	$.cookie('userStatus', $(this).children().attr("data-status"));
	location.href = "/Admin/Users?page=1";
})
$("[data-action='delete'").click(function() {
	var userId = $(this).attr("data-id");
	$(this).attr("disabled", "true");
	var thisObj = $(this);
	$.post("/Admin/DeleteUser", {
		id: userId
	}, function(data) {
		if (data.error == true) {
			toastr["warning"](data.msg);
			thisObj.removeAttr("disabled");
		} else {
			toastr["success"]("用户已删除");
			thisObj.removeAttr("disabled");
			$('[data-toggle="tooltip"]').tooltip("hide")
			thisObj.parent().parent().remove();
		}
	});
})
$("[data-action='ban'").click(function() {
	var userId = $(this).attr("data-id");
	$(this).attr("disabled", "true");
	var thisObj = $(this);
	$.post("/Admin/BanUser", {
		id: userId
	}, function(data) {
		if (data.error == 1) {
			toastr["warning"](data.msg);
			thisObj.removeAttr("disabled");
		} else {
			toastr["success"]("操作成功");
			thisObj.removeAttr("disabled");
			$('[data-toggle="tooltip"]').tooltip("hide");
			location.reload();
		}
	});
})
$("[data-type='all']").click(function(){
	$('input[type=checkbox]').prop('checked', $(this).prop('checked'));
})
$('input[type=checkbox]').click(function(){
	$("#del").show();
})
$("#delAll").click(function(){
	$("#delAll").attr("disabled", "true");
	var idGroup=[];
	$('input[data-type="mark"]:checked').each(function(){  
		idGroup.push($(this).attr("data-id"));
	});
	$.post("/Admin/DeleteUsers", {
		id: JSON.stringify(idGroup)
	}, function(data) {
		location.href = "/Admin/Users?page=1";
	});
})
$("[data-action='edit'").click(function() {
	var userId = $(this).attr("data-id");
	$(this).attr("disabled", "true");
	var thisObj = $(this);
	$.post("/Admin/GetUserInfo", {
		id: userId
	}, function(data) {
		$('[data-toggle="tooltip"]').tooltip("hide")
		thisObj.removeAttr("disabled");
		$('#editUser').modal("hide");
		$('#editUser').modal("show");
		$("#user_avatar").attr("src","/Member/Avatar/"+data.id+"/s");
		$("#id").val(data.id);
		$("#uid").val(data.id);
		$("#user_nick").val(data.user_nick);
		$("#user_email").val(data.user_email);
		$("#user_date").val(data.user_date);
		$("#used_storage").val(data.used_storage);
		$("#two_step").val(data.two_step);
		$("#user_status").val(data.user_status);
		$("#user_group").val(data.user_group);
		$("#profile"+data.profile).prop("checked","checked");
	});
})
$("#editUserForm").submit(function() {
	$("#editUserSubmit").attr("disabled", "true");
	$.post("/Admin/SaveUser", 
		$("#editUserForm").serialize()
	, function(data) {
		if (data.error == "1") {
			toastr["warning"](data.msg);
			$("#editUserSubmit").removeAttr("disabled");
		} else if (data.error == "200") {
			toastr["success"]("编辑成功");
			document.getElementById("editUserForm").reset();
			$("#editUserSubmit").removeAttr("disabled");
			$('#editUser').modal("hide");
		}else{
			toastr["warning"]("未知错误");
			$("#editUserSubmit").removeAttr("disabled");
		}
	});
	return false;
})
$("#addUserForm").submit(function() {
	$("#addUserSubmit").attr("disabled", "true");
	$.post("/Admin/AddUser", 
		$("#addUserForm").serialize()
	, function(data) {
		if (data.error == true) {
			toastr["warning"](data.msg);
			$("#addUserSubmit").removeAttr("disabled");
		} else if (data.error == false) {
			toastr["success"]("用户已添加");
			document.getElementById("addUserForm").reset();
			$("#addUserSubmit").removeAttr("disabled");
		}else{
			toastr["warning"]("未知错误");
			$("#addUserSubmit").removeAttr("disabled");
		}
	});
	return false;
})