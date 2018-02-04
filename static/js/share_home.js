		function getMemory() {
			$.get("/Member/Memory", function(data) {
				var dataObj = eval("(" + data + ")");
				if (dataObj.rate >= 100) {
					$("#memory_bar").css("width", "100%");
					$("#memory_bar").addClass("progress-bar-warning");
					toastr["error"]("您的已用容量已超过容量配额，请尽快删除多余文件或购买容量");

				} else {
					$("#memory_bar").css("width", dataObj.rate + "%");
				}
				$("#used").html(dataObj.used);
				$("#total").html(dataObj.total);
			});
		}

		window.onload = function() {
			$.material.init();
			getMemory();
			$("[href^='/Share/My']").addClass("active");
		}
		jQuery.ajaxSetup({
			cache: true
		});
		upload_load = 0;



		$(function() {
			$('[data-toggle="tooltip"]').tooltip()
		})
deleteId = 0;
		function openShare(id) {
			window.open("/s/" + id, "_blank");
		}

		function deleteShare(id) {
			deleteId = id;
			$('#deleteConfirm').modal()
		}

		function deleteConfirm() {
			$(".pro-btn").attr("disabled",true);
			$.post("/Share/Delete", {id:deleteId}, function(r){
				if(r.error){
					$(".pro-btn").removeAttr("disabled");
					toastr["error"](r.msg);
				}else{
					$(".pro-btn").removeAttr("disabled");
					toastr["success"](r.msg);
					$('#deleteConfirm').modal('hide')
					$("#"+deleteId).hide();
				}
			})
		}
		function changeType(id,type){
			changeId = id;
			$('#changeConfirm').modal()
			$("#shareType").html(type=="public"?"私密分享":"公开分享");
		}
		function changeConfirm(){
			$(".pro-btn").attr("disabled",true);
			$.post("/Share/ChangePromission", {id:changeId}, function(r){
				if(r.error){
					$(".pro-btn").removeAttr("disabled");
					toastr["error"](r.msg);
				}else{
					$(".pro-btn").removeAttr("disabled");
					toastr["success"](r.msg);
					location.reload();
					$('#changeConfirm').modal('hide')
				}
			})
		}