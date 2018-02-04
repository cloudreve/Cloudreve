            $.material.init();

            function submit() {
                $("#submit_pwd").attr("disabled",1);
                pwd = $("#inputPassword").val();
                if (pwd == "") {
                    toastr["error"]("密码不能为空");
                    $("#submit_pwd").removeAttr("disabled");
                } else {
                    $.post("/Share/chekPwd", {
                        password: pwd,
                        key: shareInfo.shareId
                    }, function(result) {
                        if (result.error) {
                            toastr["error"](result.msg);
                            $("#submit_pwd").removeAttr("disabled")
                        } else {
                            location.reload();
                        }
                    });
                }
            }
            $("#submit_pwd").click(function() {
                submit();
            });

            document.onkeyup = function(e) {
                var code = e.charCode || e.keyCode;
                if (code == 13) {
                    submit();
                }
            }