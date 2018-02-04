            $.material.init();
            function GetQueryString(name) {
                 var reg = new RegExp("(^|&)"+ name +"=([^&]*)(&|$)");
                 var r = window.location.search.substr(1).match(reg);
                 if(r!=null)return  unescape(r[2]); return null;
            }
            if(GetQueryString("type")==null){
                $("#all").addClass("active");
            }else{
                $("#hot").addClass("active");
            }