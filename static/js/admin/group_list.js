$("#dataTable_length").change(function() {
	var p1 = $(this).children('option:selected').val();
	$.cookie('pageSize', p1);
	location.href = "/Admin/GroupList?page=1";
});
$(document).ready(function() {
	var pageSize = ($.cookie('pageSize') == null) ? "10" : $.cookie('pageSize');
	$("#dataTable_length").val(pageSize);
})
$("[data-action='delete'").click(function() {
	var policyId = $(this).attr("data-id");
	if($(this).attr("data-unable") == "1"){
		toastr["warning"]("此用户组下仍有用户，请先删除这些用户");
		return false;
	}
	$(this).attr("disabled", "true");
	var thisObj = $(this);
	$.post("/Admin/DeleteGroup", {
		id: policyId
	}, function(data) {
		if (data.error == true) {
			toastr["warning"](data.msg);
			thisObj.removeAttr("disabled");
		} else {
			toastr["success"]("上传策略已删除");
			thisObj.removeAttr("disabled");
			$('[data-toggle="tooltip"]').tooltip("hide")
			thisObj.parent().parent().remove();
		}
	});
})
$("[data-action='edit'").click(function() {
	location.href = '/Admin/EditGroup/id/'+$(this).attr("data-id");
});