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

		})
		
		selected_pack = null;

		function colorChange(id,c){
			$("#"+id+"_head").animate({
  　　　		backgroundColor:c
  		},0);
			$("#"+id+"_button").animate({
  　　　		backgroundColor:c
  		},0);
		}

		for (var pack in pack_data) {

			var pack_content = '<div class="col-md-2 " ><div class="pack_container"> <a onclick="selectPack(' + pack_data[pack]["id"] + ',' + pack + ')" id="pack-id-' + pack_data[pack]["id"] + '" href="javascript:void"><div class="pack_box"><div class="box_head"><h5>' + pack_data[pack]["name"] + '</h5><div class="box_price">￥' + pack_data[pack]["price"] + '</div></div><div class="box_bottom">有效期：' + Math.ceil(pack_data[pack]["time"] / 86400) + '天</div></div></a></div></div>';
			var pack_content = '<div class="col-sm-3"><div class="card pt-item"><div class="pti-header bgm-amber" id="' + pack_data[pack]["id"] + '_head"><h2>￥' + pack_data[pack]["price"] + ' <small>| ' + Math.ceil(pack_data[pack]["time"] / 86400) + '天</small></h2><div class="ptih-title">' + pack_data[pack]["name"] + '</div> </div><div class="pti-body">' + pack_data[pack]["des"] + '</div><div class="pti-footer"><a href="javascript:void(0);" onclick="selectPack(' + pack_data[pack]["id"] + ',' + pack + ')" class="bgm-amber" id="' + pack_data[pack]["id"] + '_button"><i class="fa fa-shopping-cart"></i></a></div></div><div class="col-sm-3">';
			$("#packs").append(pack_content);
			colorChange(pack_data[pack]["id"], pack_data[pack]["color"]);
		}

		function querryLoop(id) {
			$.get("/Buy/querryStatus?id=" + id, function(result) {
				result1 = eval('(' + result + ')');
				if (result1.status == "1") {
					$(".scan").hide();
					$(".success_info").fadeIn();
					window.clearInterval(IntervalId);
				}
			});
		}
		selected_index = 0;

		function selectPack(id, pack_index) {
			$('#buy_form').modal();
			$("#group_name").html(pack_data[pack_index]["name"]);
			var count = $("#count").val();
			if (count <= 0) {
				count = 1;
			}
			$("[id^='pack-id']").children().removeClass("pack_active");
			$("#pack-id-" + id).children().addClass("pack_active");
			$(".price_num").html("￥" + intToFloat((pack_data[pack_index]["price"]) * count));
			selected_pack = id;
			selected_index = pack_index;
		}
		IntervalId = 0;
		payment_type = "jinshajiang";
		$("#buy").click(function() {
			if (selected_pack == null) {
				toastr["warning"]("请先选择一个用户组");

			} else {
				var count = parseInt($("#count").val());
				if(count>99||count<0){
					toastr["warning"]("购买时长必须在1-99之间");
				}else{
					$("#buy").attr("disabled", "true");
					$.post("/Buy/PlaceOrder", {
						action: "group",
						id: selected_pack,
						num:count
					}, function(data) {
						if (data.error == "1") {
							toastr["error"](data.msg);
							$("#buy").removeAttr("disabled");
							$('#buy_form').modal('hide');
						} else if (data.error == "200") {
							if (payment_type == "jinshajiang") {
								$("#jinshajiang").find("[name='total']").val(data.total);
								$("#jinshajiang").find("[name='showurl']").val(data.showurl);
								$("#jinshajiang").find("[name='uid']").val(data.uid);
								$("#jinshajiang").find("[name='apiid']").val(data.apiid);
								$("#jinshajiang").find("[name='apikey']").val(data.apikey);
								$("#jinshajiang").find("[name='addnum']").val(data.addnum);
								$("#jinshajiang").submit()
							}
							$("#buy").removeAttr("disabled");
						} else if (data.error == "201") {
							$(".buy_form").hide()
							$(".scan").fadeIn();
							$("#qr_code").attr("src", data.qrcode);
							IntervalId = window.setInterval(function() {
								querryLoop(data.id);
							}, 3000);
						}
					});
				}
			}
		});

		function intToFloat(val) {
			return new Number(val).toFixed(2);
		}
		$('#count').bind('input propertychange', function() {
			var count = $("#count").val();
			if (count <= 0) {
				count = 1;
			}
			if (selected_pack == null) {} else {
				$(".price_num").html("￥" + intToFloat((pack_data[selected_index]["price"]) * count));

			}
		});