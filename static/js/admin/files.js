$("#dataTable_length").change(function() {
	var p1 = $(this).children('option:selected').val();
	$.cookie('pageSize', p1);
	location.href = "/Admin/Files?page=1";
});
$(document).ready(function() {
	var pageSize = ($.cookie('pageSize') == null) ? "10" : $.cookie('pageSize');
	var policy = ($.cookie('filePolicy') == null) ? "" : $.cookie('filePolicy');
	var  searchCol = ($.cookie('searchCol') == null) ? "id" : $.cookie('searchCol');
	$("#dataTable_length").val(pageSize);
	$("#searchFrom").val($.cookie('fileSearch'));
	$("a[data-method='" + $.cookie('orderMethodFile') + "']").addClass("active");
	$("a[data-policy='" + policy + "']").addClass("active");
	$("#searchCol").val(searchCol);
	$("#searchValue").val($.cookie('searchValue'));
})

$("#searchFrom").keydown(function(e) {
	var curKey = e.which;
	if (curKey == 13) {
		$.cookie('fileSearch', $(this).val());
		location.href = "/Admin/Files?page=1";
	}
});
$("#applySearch").click(function(){
	$.cookie('searchCol', $("#searchCol").val());
	$.cookie('searchValue', $("#searchValue").val());
	location.href = "/Admin/Files?page=1";
})
$("#order").children().click(function() {
	$.cookie('orderMethodFile', $(this).children().attr("data-method"));
	location.href = "/Admin/Files?page=1";
})
$("#policy_select").children().click(function() {
	$.cookie('filePolicy', $(this).children().attr("data-policy"));
	location.href = "/Admin/Files?page=1";
})
$("[data-action='delete'").click(function() {
	var fileId = $(this).attr("data-id");
	$(this).attr("disabled", "true");
	var thisObj = $(this);
	$.post("/Admin/Delete", {
		id: fileId
	}, function(data) {
		if (data.result.success == false) {
			toastr["warning"]("删除失败");
			$(this).removeAttr("disabled");
		} else {
			toastr["success"]("文件已删除");
			$(this).removeAttr("disabled");
			$('[data-toggle="tooltip"]').tooltip("hide")
			thisObj.parent().parent().remove();
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
	var j = {};
	$('input[data-type="mark"]:checked').each(function(){  
		j.path=$(this).attr("data-path");
		j.uid=$(this).attr("data-uid");
		idGroup.push(j);
		j = {};
	});
	$.post("/Admin/DeleteMultiple", {
		id: JSON.stringify(idGroup)
	}, function(data) {
		location.href = "/Admin/Files?page=1";
	});
})
$("[data-action='download'").click(function() {
	window.open('/Admin/Download/id/'+$(this).attr("data-id"),'target','');
});
$("[data-action='info'").click(function() {
	var fileId = $(this).attr("data-id");
	$(this).attr("disabled", "true");
	var thisObj = $(this);
	$.post("/Admin/GetFileInfo", {
		id: fileId
	}, function(data) {
		$('[data-toggle="tooltip"]').tooltip("hide")
		thisObj.removeAttr("disabled");
		$('#fileInfo').modal("hide");
		$('#fileInfo').modal("show");
		$("#fileId").html(data.id);
		$("#fileName").html(data.orign_name);
		$("#fileOrigin").html(data.pre_name);
		$("#fileUID").html(data.upload_user);
		$("#fileSize").html(data.size);
		$("#filePic").html(data.pic_info);
		$("#filePolicy").html(data.policy.policy_name);
		$("#fileDir").html(data.dir);
	});
})