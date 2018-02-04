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
				$("#m_used").html(dataObj.used);
				$("#m_pack").html(dataObj.pack);
				$("#m_basic").html(dataObj.basic);
				$("#m_total").html(dataObj.total);
				$("#r1").css("width", dataObj.r1 + "%");
				$("#r2").css("width", dataObj.r2 + "%");
				$("#r3").css("width", dataObj.r3 + "%");
			});
		}

		window.onload = function() {
			$.material.init();

			getMemory();
			$("[href^='/Home/Quota']").addClass("active");
		}
		$(function() {
			$('[data-toggle="tooltip"]').tooltip()
		})