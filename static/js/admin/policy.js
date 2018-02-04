$("#dataTable_length").change(function() {
	var p1 = $(this).children('option:selected').val();
	$.cookie('pageSize', p1);
	location.href = "/Admin/PolicyList?page=1";
});
$(document).ready(function() {
	var pageSize = ($.cookie('pageSize') == null) ? "10" : $.cookie('pageSize');
	var policy = ($.cookie('policyType') == null) ? "" : $.cookie('policyType');
	$("#dataTable_length").val(pageSize);
	$("#searchFrom").val($.cookie('policySearch'));
	$("a[data-policy='" + policy + "']").addClass("active");
})
$("#searchFrom").keydown(function(e) {
	var curKey = e.which;
	if (curKey == 13) {
		$.cookie('policySearch', $(this).val());
		location.href = "/Admin/PolicyList?page=1";
	}
});
$("#policyType").children().click(function() {
	$.cookie('policyType', $(this).children().attr("data-policy"));
	location.href = "/Admin/policyList?page=1";
})
$("[data-action='delete'").click(function() {
	var policyId = $(this).attr("data-id");
	if($(this).attr("data-unable") == "1"){
		toastr["warning"]("此上传方案下仍有文件，请先删除这些文件。");
		return false;
	}
	$(this).attr("disabled", "true");
	var thisObj = $(this);
	$.post("/Admin/DeletePolicy", {
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
	location.href = '/Admin/EditPolicy/id/'+$(this).attr("data-id");
});