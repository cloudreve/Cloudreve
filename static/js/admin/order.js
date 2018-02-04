$("#dataTable_length").change(function() {
	var p1 = $(this).children('option:selected').val();
	$.cookie('pageSize', p1);
	location.href = "/Admin/OrderList?page=1";
});
$(document).ready(function() {
	var pageSize = ($.cookie('pageSize') == null) ? "10" : $.cookie('pageSize');
	var orderType = ($.cookie('orderType') == null) ? "" : $.cookie('orderType');
	var orderStatus = ($.cookie('orderStatus') == null) ? "" : $.cookie('orderStatus');
	$("#dataTable_length").val(pageSize);
	$("a[data-type='" + orderType + "']").addClass("active");
	$("a[data-status='" + orderStatus + "']").addClass("active");
})
$("#orderType").children().click(function() {
	$.cookie('orderType', $(this).children().attr("data-type"));
	location.href = "/Admin/OrderList?page=1";
})
$("#orderStatus").children().click(function() {
	$.cookie('orderStatus', $(this).children().attr("data-status"));
	location.href = "/Admin/OrderList?page=1";
})
$("[data-action='delete'").click(function() {
	var orderId = $(this).attr("data-id");
	$(this).attr("disabled", "true");
	var thisObj = $(this);
	$.post("/Admin/DeleteOrder", {
		id: orderId
	}, function(data) {
		if (data.error == true) {
			toastr["warning"](data.msg);
			thisObj.removeAttr("disabled");
		} else {
			toastr["success"]("订单已删除");
			thisObj.removeAttr("disabled");
			$('[data-toggle="tooltip"]').tooltip("hide")
			thisObj.parent().parent().remove();
		}
	});
})