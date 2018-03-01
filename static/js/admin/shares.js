$("#dataTable_length").change(function() {
	var p1 = $(this).children('option:selected').val();
	$.cookie('pageSize', p1);
	location.href = "/Admin/Shares?page=1";
});
$("[data-type='all']").click(function(){
	$('input[type=checkbox]').prop('checked', $(this).prop('checked'));
})
$('input[type=checkbox]').click(function(){
	$("#del").show();
})
$(document).ready(function() {
	var pageSize = ($.cookie('pageSize') == null) ? "10" : $.cookie('pageSize');
	var shareType = ($.cookie('shareType') == null) ? "" : $.cookie('shareType');
	$("#dataTable_length").val(pageSize);
	$("#searchFrom").val($.cookie('shareSearch'));
	$("a[data-type='" + shareType + "']").addClass("active");
	$("a[data-method='" + $.cookie('orderMethodShare') + "']").addClass("active");
})
$("#searchFrom").keydown(function(e) {
	var curKey = e.which;
	if (curKey == 13) {
		$.cookie('shareSearch', $(this).val());
		location.href = "/Admin/Shares?page=1";
	}
});
$("#group").children().click(function() {
	$.cookie('shareType', $(this).children().attr("data-type"));
	location.href = "/Admin/Shares?page=1";
})
$("#order").children().click(function() {
	$.cookie('orderMethodShare', $(this).children().attr("data-method"));
	location.href = "/Admin/Shares?page=1";
})
$("[data-action='delete']").click(function() {
	var shareId = $(this).attr("data-id");
	$(this).attr("disabled", "true");
	var thisObj = $(this);
	$.post("/Admin/DeleteShare", {
		id: shareId
	}, function(data) {
		if (data.error == true) {
			toastr["warning"]("删除失败");
			$(this).removeAttr("disabled");
		} else {
			toastr["success"]("分享已删除");
			$(this).removeAttr("disabled");
			$('[data-toggle="tooltip"]').tooltip("hide")
			thisObj.parent().parent().remove();
		}
	});
})
$("[data-action='change']").click(function() {
	var shareId = $(this).attr("data-id");
	$(this).attr("disabled", "true");
	var thisObj = $(this);
	$.post("/Admin/ChangeShareType", {
		id: shareId
	}, function(data) {
		if (data.error == true) {
			toastr["warning"]("切换失败");
			thisObj.removeAttr("disabled");
		} else {
			toastr["success"](data.msg);
			thisObj.removeAttr("disabled");
		}
	});
})
$("#delAll").click(function(){
	$("#delAll").attr("disabled", "true");
	var idGroup=[];
	var j = {};
	$('input[data-type="mark"]:checked').each(function(){  
		path=$(this).attr("data-id");
		idGroup.push(path);
	});
	$.post("/Admin/DeleteShareMultiple", {
		id: JSON.stringify(idGroup)
	}, function(data) {
		location.href = "/Admin/Shares?page=1";
	});
})
