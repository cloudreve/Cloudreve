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
		}
		$(function() {
			$('[data-toggle="tooltip"]').tooltip()
		})
		$("#avatar_file").on("change", function() {
			ajaxFileUpload();
		})

		function ajaxFileUpload() {
			$("#upload-text").html("正在上传...");
			$.ajaxFileUpload({
				url: '/Member/SaveAvatar', //用于文件上传的服务器端请求地址
				secureuri: false,
				fileElementId: 'uploadAvatar', //文件上传域的ID
				dataType: 'json', //返回值类型 一般设置为json
				error: function(data) //服务器响应失败处理函数
					{
						data = eval("(" + data.responseText + ")");
						if (data.result == "success") {
							location.reload();
						} else {
							toastr["warning"](data.msg);
							$("#avatar_file").on("change", function() {
								ajaxFileUpload();
							})
							$("#upload-text").html("上传头像");
						}
					}
			})
			return false;
		};

		$("#saveNick").click(function() {
			var newNick = $("#nick").val();
			$("#saveNick").attr("disabled", "true");
			$.post("/Member/Nick", {
				nick: newNick
			}, function(data) {
				if (data.error == "1") {
					toastr["warning"](data.msg);
					$("#saveNick").removeAttr("disabled");
				} else if (data.error == "200") {
					location.reload();
				}
			});
		})

		$("#homePage").change(function() {
			if ($(this).prop("checked")) {
				postData = "true";
			} else {
				postData = "false";
			}
			$.post("/Member/HomePage", {
				status: postData
			}, function(data) {
				if (data.error == "1") {
					toastr["warning"](data.msg);
				} else if (data.error == "200") {
					toastr["success"](data.msg);
				}
			});
		})

		$("#twoStep").click(function(){
			$("#two_step_modal").modal();
			$("#qrcode").attr("src","/Member/EnableTwoFactor");
		})

		$("#setWebdavPwd").click(function(){
			$("#set_webdav_pwd").modal();
		})

		$("#confirm").click(function(){
			$vCode = $("#vCode").val();
			$("#confirm").attr("disabled", "true");
			$.post("/Member/TwoFactorConfirm", {
				code: $vCode
			}, function(data) {
				if (data.error == "1") {
					$("#confirm").removeAttr("disabled");
					toastr["warning"](data.msg);
				} else if (data.error == "200") {
					toastr["success"](data.msg);
					location.reload();
				}
			});
		})

		$("#confirmWebdav").click(function(){
			pwd = $("#webdav_pwd").val();
			$("#confirmWebdav").attr("disabled", "true");
			$.post("/Member/setWebdavPwd", {
				pwd: pwd
			}, function(data) {
				if (data.error == "1") {
					$("#confirmWebdav").removeAttr("disabled");
					toastr["warning"](data.msg);
				} else if (data.error == "200") {
					toastr["success"](data.msg);
					$("#confirmWebdav").removeAttr("disabled");
					$("#set_webdav_pwd").modal('hide');
				}
			});
		})

		$("#savePwd").click(function(){
			$("#savePwd").attr("disabled","true");
			var pwdOrigin=$("#passOrigin").val();
			var pwdNew=$("#passNew").val();
			var pwdNewRepet=$("#passNewRepet").val();
			if(pwdNew != pwdNewRepet){
				toastr["warning"]("两次密码输入不一致");
				$("#savePwd").removeAttr("disabled");
				return 0;
			}
			$.post("/Member/ChangePwd", {origin:pwdOrigin,new:pwdNew}, function(data) {
				if (data.error == "1") {
					$("#savePwd").removeAttr("disabled");
					toastr["warning"](data.msg);
				}else if (data.error == "200") {
					toastr["success"](data.msg);
					location.reload();
				}
			})
		})

		$("#useGravatar").click(function(){
			$("#useGravatar").attr("disabled", "true");
			$.post("/Member/SetGravatar", {
				"t":"confirmed"
			}, function(data) {
				location.reload();
			});
		})