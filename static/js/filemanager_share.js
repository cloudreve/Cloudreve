!
function(e, r, n) {
    "use strict";
    r.module("FileManagerApp", ["pascalprecht.translate"]),
        n(e.document).on("shown.bs.modal", ".modal",
            function() {
                e.setTimeout(function() {
                    n("[autofocus]", this).focus()
                }.bind(this), 100)
            }),
        n(e.document).on("click",
            function() {
                n("#context-menu").hide();
                mobileBind();
            }),
        n(e.document).on("contextmenu", '.main-navigation .table-files tr.item-list:has("td"), .item-list',
            function(r) {
                var i = n("#context-menu");
                r.pageX >= e.innerWidth - i.width() && (r.pageX -= i.width()),
                    r.pageY >= e.innerHeight - i.height() && (r.pageY -= i.height()),
                    i.hide().css({
                        left: r.pageX,
                        top: r.pageY
                    }).appendTo("body").show(),
                    r.preventDefault()
            }),
        Array.prototype.find || (Array.prototype.find = function(e) {
            if (null == this) throw new TypeError("Array.prototype.find called on null or undefined");
            if ("function" != typeof e) throw new TypeError("predicate must be a function");
            for (var r, n = Object(this), i = n.length >>> 0, a = arguments[1], t = 0; i > t; t++)
                if (r = n[t], e.call(a, r, t, n)) return r
        })
}(window, angular, jQuery),
function(e, r) {
    "use strict";
    e.module("FileManagerApp").controller("FileManagerCtrl", ["$scope", "$rootScope", "$window", "$translate", "fileManagerConfig", "item", "fileNavigator", "apiMiddleware",
        function(e, n, i, a, t, o, s, l) {
            var d = i.localStorage;
            e.config = t,
                e.reverse = !1,
                e.predicate = ["model.type", "model.name"],
                e.order = function(r) {
                    e.reverse = e.predicate[1] === r ? !e.reverse : !1,
                        e.predicate[1] = r
                },
                e.query = "",
                e.fileNavigator = new s,
                e.apiMiddleware = new l,
                e.uploadFileList = [],
                e.viewTemplate = d.getItem("viewTemplate") || "main-icons.html",
                e.fileList = [],
                e.temps = [],
                e.$watch("temps",
                    function() {
                        e.singleSelection() ? e.temp = e.singleSelection() : (e.temp = new o({
                                rights: 644
                            }), e.temp.multiple = !0),
                            e.temp.revert(),
                             $.material.init();
                    }),
                e.fileNavigator.onRefresh = function() {
                    e.temps = [],
                        e.query = "",
                        n.selectedModalPath = e.fileNavigator.currentPath,
                        $.cookie('path_tmp', n.selectedModalPath); 
                },
                e.setTemplate = function(r) {
                    d.setItem("viewTemplate", r),
                        e.viewTemplate = r
                },
                e.changeLanguage = function(e) {
                    return e ? (d.setItem("language", e), a.use(e)) : void a.use(d.getItem("language") || t.defaultLang)
                },
                e.isSelected = function(r) {
                    return -1 !== e.temps.indexOf(r)
                },
                e.selectOrUnselect = function(r, n) {
                    var i = e.temps.indexOf(r),
                        a = n && 3 == n.which;
                    if (n && n.target.hasAttribute("prevent")) return void(e.temps = []);
                    if (!(!r || a && e.isSelected(r))) {
                        if (n && n.shiftKey && !a) {
                            var t = e.fileList,
                                o = t.indexOf(r),
                                s = e.temps[0],
                                l = t.indexOf(s),
                                d = void 0;
                            if (s && t.indexOf(s) < o) {
                                for (e.temps = []; o >= l;) d = t[l], !e.isSelected(d) && e.temps.push(d),
                                    l++;
                                return
                            }
                            if (s && t.indexOf(s) > o) {
                                for (e.temps = []; l >= o;) d = t[l], !e.isSelected(d) && e.temps.push(d),
                                    l--;
                                return
                            }
}                return n && !a && (n.ctrlKey || n.metaKey) ? void(e.isSelected(r) ? e.temps.splice(i, 1) : e.temps.push(r)) : void(e.temps = [r])
                    }
                },
                e.singleSelection = function() {
                    return 1 === e.temps.length && e.temps[0]
                },
                e.totalSelecteds = function() {
                    return {
                        total: e.temps.length
                    }
                },
                e.selectionHas = function(r) {
                    return e.temps.find(function(e) {
                        return e && e.model.type === r
                    })
                },
                e.prepareNewFolder = function() {
                    var r = new o(null, e.fileNavigator.currentPath);
                    return e.temps = [r],
                        r
                },
                e.smartClick = function(r) {
                    var n = e.config.allowedActions.pickFiles;
                    if (r.isFolder()) return e.fileNavigator.folderClick(r);
                    if ("function" == typeof e.config.pickCallback && n) {
                        var i = e.config.pickCallback(r.model);
                        if (i === !0) return
                    }
                    return r.isImage() ? e.config.previewImagesInModal ? e.openImagePreview(r) : e.apiMiddleware.download(r, !0) : r.isEditable() ? e.openEditItem(r) : void 0
                },
                e.openImagePreview = function() {
                    var r = e.singleSelection();
                    if(r.model.pic==""){
                    }else{
                        t =e.apiMiddleware.listPic(r);
                       loadPreview(t);

                    }
  
      

    
            
                },
                e.openGetSource = function() {
                    var r = e.singleSelection();
                    e.apiMiddleware.apiHandler.inprocess = !0,
                        e.modal("getsource", null, !0).find("#source-target").attr("tmp", e.apiMiddleware.getsource(r)).unbind("load error");
                        if(cliLoad !=1){
                        var clipboard = new Clipboard('.btn-copy');
                        cliLoad =1;
                        clipboard.on('success', function(e) {
                            toastr["success"]("复制成功");
                        })}
                        e.apiMiddleware.apiHandler.inprocess = !1;
                      
                         
                },
                e.openVideoPreview = function() {
                    var r = e.singleSelection();
                    e.apiMiddleware.apiHandler.inprocess = !1,
                        e.modal("videopreview", null, !0);
                        loadDPlayer(e.apiMiddleware.preview(r));
                },
          e.openAudioPreview = function() {
                    var r = e.singleSelection();
                    e.apiMiddleware.apiHandler.inprocess = !1,
                        e.modal("audiopreview", null, !0).find("#audiopreview-target").attr("src", e.apiMiddleware.preview(r)).unbind("load error").on("load error",
                            function() {
                                e.apiMiddleware.apiHandler.inprocess = !1,
                                    e.$apply()
                            })
                },
                e.openEditItem = function() {
                    var r = e.singleSelection();
                    e.apiMiddleware.getContent(r).then(function(e) {
                            r.tempModel.content = r.model.content = e.result
                        }),
                        e.modal("edit")
                },
                e.modal = function(n, i, a) {
                    var t = r("#" + n);
                    return t.modal(i ? "hide" : "show"),
                        e.apiMiddleware.apiHandler.error = "",
                        e.apiMiddleware.apiHandler.asyncSuccess = !1,
                        a ? t : !0
                },
                e.modalWithPathSelector = function(r) {
                    return n.selectedModalPath = e.fileNavigator.currentPath,
                        e.modal(r)
                },
                e.isInThisPath = function(r) {
                    var n = e.fileNavigator.currentPath.join("/") + "/";
                    return -1 !== n.indexOf(r + "/")
                },
                e.edit = function() {
                    e.apiMiddleware.edit(e.singleSelection()).then(function() {
                        e.modal("edit", !0)
                    })
                },
                e.changePermissions = function() {
                    e.apiMiddleware.changePermissions(e.temps, e.temp).then(function() {
                        e.fileNavigator.refresh(),
                            e.modal("changepermissions", !0)
                    })
                },
                e.download = function() {
                    var r = e.singleSelection();
                    if (!e.selectionHas("dir")) return r ? e.apiMiddleware.download(r) : e.apiMiddleware.downloadMultiple(e.temps)
                },
                e.copy = function() {
                    var r = e.singleSelection();
                    if (r) {
                        var i = r.tempModel.name.trim(),
                            t = e.fileNavigator.fileNameExists(i);
                        if (t && c(r)) return e.apiMiddleware.apiHandler.error = a.instant("error_invalid_filename"), !1;
                        if (!i) return e.apiMiddleware.apiHandler.error = a.instant("error_invalid_filename"), !1
                    }
                    e.apiMiddleware.copy(e.temps, n.selectedModalPath).then(function() {
                        e.fileNavigator.refresh(),
                            e.modal("copy", !0)
                    })
                },
                e.compress = function() {
                    var r = e.temp.tempModel.name.trim(),
                        i = e.fileNavigator.fileNameExists(r);
                    return i && c(e.temp) ? (e.apiMiddleware.apiHandler.error = a.instant("error_invalid_filename"), !1) : r ? void e.apiMiddleware.compress(e.temps, r, n.selectedModalPath).then(function() {
                            return e.fileNavigator.refresh(),
                                e.config.compressAsync ? void(e.apiMiddleware.apiHandler.asyncSuccess = !0) : e.modal("compress", !0)
                        },
                        function() {
                            e.apiMiddleware.apiHandler.asyncSuccess = !1
                        }) : (e.apiMiddleware.apiHandler.error = a.instant("error_invalid_filename"), !1)
                },
                e.extract = function() {
                    var r = e.temp,
                        i = e.temp.tempModel.name.trim(),
                        t = e.fileNavigator.fileNameExists(i);
                    return t && c(e.temp) ? (e.apiMiddleware.apiHandler.error = a.instant("error_invalid_filename"), !1) : i ? void e.apiMiddleware.extract(r, i, n.selectedModalPath).then(function() {
                            return e.fileNavigator.refresh(),
                                e.config.extractAsync ? void(e.apiMiddleware.apiHandler.asyncSuccess = !0) : e.modal("extract", !0)
                        },
                        function() {
                            e.apiMiddleware.apiHandler.asyncSuccess = !1
                        }) : (e.apiMiddleware.apiHandler.error = a.instant("error_invalid_filename"), !1)
                },
                e.remove = function() {
                        var dirList= new Array();
                        var fileList = new Array();
                        for(var x in e.temps){
                            if (e.temps[x].model.type == "dir"){
                                dirList.push(e.temps[x]);

                            }else{
                                fileList.push(e.temps[x]);
                            }
                            
                        }
                        //console.log(dirList);
                        
                    e.apiMiddleware.remove(fileList,dirList).then(function() {
                        e.fileNavigator.refresh(),
                            e.modal("remove", !0)
                            getMemory();

                    })
                },
                e.move = function() {
                     var dirList= new Array();
                        var fileList = new Array();
                        for(var x in e.temps){
                            if (e.temps[x].model.type == "dir"){
                                dirList.push(e.temps[x]);

                            }else{
                                fileList.push(e.temps[x]);
                            }
                            
                        }
                    var r = e.singleSelection() || e.temps[0];
                    return r && c(r) ? (e.apiMiddleware.apiHandler.error = a.instant("error_cannot_move_same_path"), !1) : void e.apiMiddleware.move(fileList,dirList, n.selectedModalPath).then(function() {
                        e.fileNavigator.refresh(),
                            e.modal("move", !0)
                    })
                },
                e.rename = function() {
                    var r = e.singleSelection(),
                        n = r.tempModel.name,
                        i = r.tempModel.path.join("") === r.model.path.join("");
                    return !n || i && e.fileNavigator.fileNameExists(n) ? (e.apiMiddleware.apiHandler.error = a.instant("error_invalid_filename"), !1) : void e.apiMiddleware.rename(r).then(function() {
                        e.fileNavigator.refresh(),
                            e.modal("rename", !0)
                    })
                },
                e.sharePublic = function() {
                    var r = e.singleSelection(),
                        n = r.tempModel.name,
                        i = r.tempModel.path.join("") === r.model.path.join("");
                    return void e.apiMiddleware.sharep(r).then(function(ee) {
                    
                            e.modal("share", !0);
                            //console.log(r.model.name);
                            
                            e.modal("share_result",null, !0).find("#share-public-target").attr("value",ee.result);
                             document.getElementById("file_name").innerHTML = r.model.name
                            if(cliLoad !=1){
                        var clipboard = new Clipboard('.btn-copy');
                        cliLoad =1;
                        clipboard.on('success', function(e) {
                            toastr["success"]("复制成功");
                        })}
                        
                            
                    })
                },
                e.shareSecret = function() {
                    var r = e.singleSelection(),
                        n = r.tempModel.name,
                        i = r.tempModel.path.join("") === r.model.path.join("");
                    return void e.apiMiddleware.sharec(r).then(function(ee) {
                    
                            e.modal("share", !0);
                            //console.log(r.model.name);
                            
                            e.modal("share_result",null, !0).find("#share-public-target").attr("value",ee.result);
                             document.getElementById("file_name").innerHTML = r.model.name
                            if(cliLoad !=1){
                        var clipboard = new Clipboard('.btn-copy');
                        cliLoad =1;
                        clipboard.on('success', function(e) {
                            toastr["success"]("复制成功");
                        })}
                        
                            
                    })
                },
                e.createFolder = function() {
                    var r = e.singleSelection(),
                        n = r.tempModel.name;
                    return !n || e.fileNavigator.fileNameExists(n) ? e.apiMiddleware.apiHandler.error = a.instant("error_invalid_filename") : void e.apiMiddleware.createFolder(r).then(function() {
                        e.fileNavigator.refresh(),
                            e.modal("newfolder", !0)
                    })
                },
                /* hahahahahahahaha */
                e.addForUpload = function(r) {
                    e.uploadFileList = e.uploadFileList.concat(r),
                        e.modal("uploadfile")
                },
                e.removeFromUpload = function(r) {
                    e.uploadFileList.splice(r, 1)
                },
                e.uploadFiles = function() {
                    e.apiMiddleware.upload(e.uploadFileList, e.fileNavigator.currentPath).then(function() {
                            e.fileNavigator.refresh(),
                                e.uploadFileList = [],
                                e.modal("uploadfile", !0)
                        },
                        function(r) {
                            var n = r.result && r.result.error || a.instant("error_uploading_files");
                            e.apiMiddleware.apiHandler.error = n
                        })
                };
            var c = function(e) {
                    var r = n.selectedModalPath.join(""),
                        i = e && e.model.path.join("");
                    return i === r
                },
                p = function(e) {
                    var r = i.location.search.substr(1).split("&").filter(function(r) {
                        return e === r.split("=")[0]
                    });
                    return r[0] && r[0].split("=")[1] || void 0
                };
            e.changeLanguage(p("lang")),
                e.isWindows = "Windows" === p("server"),
                e.fileNavigator.refresh()
        }
    ])
}(angular, jQuery),
function(e) {
    "use strict";
    e.module("FileManagerApp").controller("ModalFileManagerCtrl", ["$scope", "$rootScope", "fileNavigator",
        function(e, r, n) {
            e.reverse = !1,
                e.predicate = ["model.type", "model.name"],
                e.fileNavigator = new n,
                r.selectedModalPath = [],
                e.order = function(r) {
                    e.reverse = e.predicate[1] === r ? !e.reverse : !1,
                        e.predicate[1] = r
                },
                e.select = function(n) {
                    r.selectedModalPath = n.model.fullPath().split("/").filter(Boolean),
                        e.modal("selector", !0)
                },
                e.selectCurrent = function() {
                    r.selectedModalPath = e.fileNavigator.currentPath,
                        e.modal("selector", !0)
                },
                e.selectedFilesAreChildOfPath = function(r) {
                    var n = r.model.fullPath();
                    return e.temps.find(function(e) {
                        var r = e.model.fullPath();
                        return n == r ? !0 : void 0
                    })
                },
                r.openNavigator = function(r) {
                    e.fileNavigator.currentPath = r,
                        e.fileNavigator.refresh(),
                        e.modal("selector")

                },
                r.getSelectedPath = function() {
                    var n = r.selectedModalPath.filter(Boolean),
                        i = "/" + n.join("/");
                    return e.singleSelection() && !e.singleSelection().isFolder() && (i += "/" + e.singleSelection().tempModel.name),
                        i.replace(/\/\//, "/")
                }
        }
    ])
}(angular),
function(e) {
    "use strict";
    var r = e.module("FileManagerApp");
    r.directive("angularFilemanager", ["$parse", "fileManagerConfig",
            function(e, r) {
                return {
                    restrict: "EA",
                    templateUrl: r.tplPath + "/main.html"
                }
            }
        ]),
        r.directive("ngFile", ["$parse",
            function(e) {
                return {
                    restrict: "A",
                    link: function(r, n, i) {
                        var a = e(i.ngFile),
                            t = a.assign;
                        n.bind("change",
                            function() {
                                r.$apply(function() {
                                    t(r, n[0].files)
                                })
                            })
                    }
                }
            }
        ]),
        r.directive("ngRightClick", ["$parse",
            function(e) {
                return function(r, n, i) {
                    var a = e(i.ngRightClick);
                    n.bind("contextmenu",
                        function(e) {
                            r.$apply(function() {
                                e.preventDefault(),
                                    a(r, {
                                        $event: e
                                    })
                            })
                        })
                }
            }
        ])
}(angular),
function(e) {
    "use strict";
    e.module("FileManagerApp").service("chmod",
        function() {
            var e = function(e) {
                if (this.owner = this.getRwxObj(), this.group = this.getRwxObj(), this.others = this.getRwxObj(), e) {
                    var r = isNaN(e) ? this.convertfromCode(e) : this.convertfromOctal(e);
                    if (!r) throw new Error("Invalid chmod input data (%s)".replace("%s", e));
                    this.owner = r.owner,
                        this.group = r.group,
                        this.others = r.others
                }
            };
            return e.prototype.toOctal = function(e, r) {
                    var n = [];
                    return ["owner", "group", "others"].forEach(function(e, r) {
                            n[r] = this[e].read && this.octalValues.read || 0,
                                n[r] += this[e].write && this.octalValues.write || 0,
                                n[r] += this[e].exec && this.octalValues.exec || 0
                        }.bind(this)),
                        (e || "") + n.join("") + (r || "")
                },
                e.prototype.toCode = function(e, r) {
                    var n = [];
                    return ["owner", "group", "others"].forEach(function(e, r) {
                            n[r] = this[e].read && this.codeValues.read || "-",
                                n[r] += this[e].write && this.codeValues.write || "-",
                                n[r] += this[e].exec && this.codeValues.exec || "-"
                        }.bind(this)),
                        (e || "") + n.join("") + (r || "")
                },
                e.prototype.getRwxObj = function() {
                    return {
                        read: !1,
                        write: !1,
                        exec: !1
                    }
                },
                e.prototype.octalValues = {
                    read: 4,
                    write: 2,
                    exec: 1
                },
                e.prototype.codeValues = {
                    read: "r",
                    write: "w",
                    exec: "x"
                },
                e.prototype.convertfromCode = function(e) {
                    if (e = ("" + e).replace(/\s/g, ""), e = 10 === e.length ? e.substr(1) : e, /^[-rwxts]{9}$/.test(e)) {
                        var r = [],
                            n = e.match(/.{1,3}/g);
                        for (var i in n) {
                            var a = this.getRwxObj();
                            a.read = /r/.test(n[i]),
                                a.write = /w/.test(n[i]),
                                a.exec = /x|t/.test(n[i]),
                                r.push(a)
                        }
                        return {
                            owner: r[0],
                            group: r[1],
                            others: r[2]
                        }
                    }
                },
                e.prototype.convertfromOctal = function(e) {
                    if (e = ("" + e).replace(/\s/g, ""), e = 4 === e.length ? e.substr(1) : e, /^[0-7]{3}$/.test(e)) {
                        var r = [],
                            n = e.match(/.{1}/g);
                        for (var i in n) {
                            var a = this.getRwxObj();
                            a.read = /[4567]/.test(n[i]),
                                a.write = /[2367]/.test(n[i]),
                                a.exec = /[1357]/.test(n[i]),
                                r.push(a)
                        }
                        return {
                            owner: r[0],
                            group: r[1],
                            others: r[2]
                        }
                    }
                },
                e
        })
}(angular),
function(e) {
    "use strict";
    e.module("FileManagerApp").factory("item", ["fileManagerConfig", "chmod",
        function(r, n) {
            var i = function(r, i) {
                function a(e) {
                    var r = (e || "").toString().split(/[- :]/);
                    return new Date(r[0], r[1] - 1, r[2], r[3], r[4], r[5])
                }
                var t = {
                    name: r && r.name || "",
                    name2: r && r.name2 || "",
                    path: i || [],
                    type: r && r.type || "file",
                    size: r && parseInt(r.size || 0),
                    date: a(r && r.date),
                    perms: new n(r && r.rights),
                    content: r && r.content || "",
                    fileId: r && r.id || '',
                    pic: r && r.pic || '',
                    recursive: !1,
                    fullPath: function() {
                        var e = this.path.filter(Boolean);
                        return ("/" + e.join("/") + "/" + this.name).replace(/\/\//, "/")
                    },
                    encodedPath:function(){
                        return encodeURI(this.fullPath())
                    }
                };
                this.error = "",
                    this.processing = !1,
                    this.model = e.copy(t),
                    this.tempModel = e.copy(t)
            };
            return i.prototype.update = function() {
                    e.extend(this.model, e.copy(this.tempModel))
                },
                i.prototype.revert = function() {
                    e.extend(this.tempModel, e.copy(this.model)),
                        this.error = ""
                },
                i.prototype.isFolder = function() {
                    return "dir" === this.model.type
                },
                i.prototype.isEditable = function() {
                    return !this.isFolder() && r.isEditableFilePattern.test(this.model.name)
                },
                i.prototype.isImage = function() {
                    return r.isImageFilePattern.test(this.model.name)
                },
                i.prototype.getShareKey = function() {
                    return shareInfo.shareId
                },
                i.prototype.isVideo = function() {
                    return r.isVideoFilePattern.test(this.model.name)
                },

                i.prototype.isAudio = function() {
                    return r.isAudioFilePattern.test(this.model.name)
                },
                i.prototype.isCompressible = function() {
                    return this.isFolder()
                },
                i.prototype.isExtractable = function() {
                    return !this.isFolder() && r.isExtractableFilePattern.test(this.model.name)
                },
                i.prototype.isSelectable = function() {
                    return this.isFolder() && r.allowedActions.pickFolders || !this.isFolder() && r.allowedActions.pickFiles
                },
                i
        }
    ])
}(angular),
function(e) {
    "use strict";
    var r = e.module("FileManagerApp");
    r.filter("strLimit", ["$filter",
            function(e) {
                return function(r, n, i) {
                    function subString(str, len, hasDot) {
            var newLength = 0;
            var newStr = "";
            var chineseRegex = /[^\x00-\xff]/g;
            var singleChar = "";
            var strLength = str.replace(chineseRegex, "**").length;
            for (var i = 0; i < strLength; i++) {
                singleChar = str.charAt(i).toString();
                if (singleChar.match(chineseRegex) != null) {
                    newLength += 2;
                }
                else {
                    newLength++;
                }
                if (newLength > len) {
                    break;
                }
                newStr += singleChar;
            }

            if (hasDot && strLength > len) {
                newStr += "...";
            }
            return newStr;
        }

                    return r.length <= n ? r : subString(r, n) + (i || "...")
                }
            }
        ]),
        r.filter("fileExtension", ["$filter",
            function(e) {
                return function(r) {
                    return /\./.test(r) && e("strLimit")(r.split(".").pop(), 3, "..") || ""
                }
            }
        ]),
        r.filter("formatDate", ["$filter",
            function() {
                return function(e) {
                    return e instanceof Date ? e.toISOString().substring(0, 19).replace("T", " ") : (e.toLocaleString || e.toString).apply(e)
                }
            }
        ]),
        r.filter("humanReadableFileSize", ["$filter", "fileManagerConfig",
            function(e, r) {
                var n = [" kB", " MB", " GB", " TB", "PB", "EB", "ZB", "YB"],
                    i = ["KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"];
                return function(e) {
                    var a = -1,
                        t = e;
                    do t /= 1024,
                        a++;
                    while (t > 1024);
                    var o = r.useBinarySizePrefixes ? i[a] : n[a];
                    return Math.max(t, .1).toFixed(1) + " " + o
                }
            }
        ])
}(angular),
function(e) {
    "use strict";
    e.module("FileManagerApp").provider("fileManagerConfig",
        function() {
            var r = {
                appName: "angular-filemanager v1.5",
                defaultLang: "zh_cn",
                listUrl: "/Share/ListFile/"+shareInfo.shareId,
                previewUrl:"/Share/Preview/"+shareInfo.shareId,
                renameUrl: "/File/Rename",
                shareUrl:"/File/Share",
                copyUrl: "bridges/php/handler.php",
                moveUrl: "/File/Move",
                removeUrl: "/File/Delete",
                editUrl: "/File/Edit",
                getContentUrl: "/File/Content",
                createFolderUrl: "/File/createFolder",
                downloadFileUrl: "/Share/Download/"+shareInfo.shareId,
                downloadMultipleUrl: "bridges/php/handler.php",
                compressUrl: "bridges/php/handler.php",
                extractUrl: "bridges/php/handler.php",
                permissionsUrl: "bridges/php/handler.php",
                sourceUrl:"/File/gerSource",
                basePath: "/",
                searchForm: !0,
                sidebar: !0,
                breadcrumb: !0,
                allowedActions: {
                    shareFile:!1,
                    getSource:!1,
                    rename: !1,
                    move: !1,
                    copy: !1,
                    edit: !1,
                    changePermissions: !1,
                    compress: !1,
                    compressChooseName: !1,
                    extract: !1,
                    download: !0,
                    downloadMultiple: !1,
                    preview: !0,
                    remove: !1,
                    createFolder: !1,
                    pickFiles: !1,
                    pickFolders: !1
                },
                multipleDownloadFileName: "angular-filemanager.zip",
                filterFileExtensions: [],
                showExtensionIcons: !0,
                showSizeForDirectories: !1,
                useBinarySizePrefixes: !1,
                downloadFilesByAjax: !0,
                previewImagesInModal: !0,
                enablePermissionsRecursive: !0,
                compressAsync: !1,
                extractAsync: !1,
                pickCallback: false,
                isEditableFilePattern: /\.(txt|diff?|patch|svg|asc|cnf|cfg|conf|html?|.html|cfm|cgi|aspx?|ini|pl|py|md|css|cs|js|jsp|log|htaccess|htpasswd|gitignore|gitattributes|env|json|atom|eml|rss|markdown|sql|xml|xslt?|sh|rb|as|bat|cmd|cob|for|ftn|frm|frx|inc|lisp|scm|coffee|php[3-6]?|java|c|cbl|go|h|scala|vb|tmpl|lock|go|yml|yaml|tsv|lst)$/i,
                isImageFilePattern: /\.(jpe?g|gif|bmp|png|svg|tiff?)$/i,
                isVideoFilePattern:/\.(mp4|flv|avi|tff?)$/i,
                isAudioFilePattern:/\.(mp3|wav|ogg?)$/i,
                isExtractableFilePattern: /\.(gz|tar|rar|g?zip)$/i,
                tplPath: "src/templates"
            };
            return {
                $get: function() {
                    return r
                },
                set: function(n) {
                    e.extend(r, n)
                }
            }
        })
}(angular),
function(e) {
    "use strict";
    e.module("FileManagerApp").config(["$translateProvider",
        function(e) {
            e.useSanitizeValueStrategy(null),
                e.translations("zh_cn", {
                    filemanager: shareInfo.dirName,
                    language: "语言",
                    english: "英语",
                    spanish: "西班牙语",
                    portuguese: "葡萄牙语",
                    french: "法语",
                    german: "德语",
                    hebrew: "希伯来语",
                    italian: "意大利",
                    slovak: "斯洛伐克语",
                    chinese_tw: "正体中文",
                    chinese_cn: "简体中文",
                    russian: "俄語",
                    ukrainian: "烏克蘭",
                    turkish: "土耳其",
                    persian: "波斯語",
                    polish: "波兰语",
                    confirm: "确定",
                    cancel: "取消",
                    close: "关闭",
                    upload_files: "上传文件",
                    files_will_uploaded_to: "文件将上传到",
                    select_files: "选择文件",
                    uploading: "上传中",
                    permissions: "权限",
                    select_destination_folder: "选择目标文件",
                    source: "源自",
                    destination: "目的地",
                    copy_file: "复制文件",
                    sure_to_delete: "确定要删除？",
                    change_name_move: "改名或移动？",
                    enter_new_name_for: "输入新的名称",
                    extract_item: "解压",
                    extraction_started: "解压已经在后台开始",
                    compression_started: "压缩已经在后台开始",
                    enter_folder_name_for_extraction: "输入解压的目标文件夹",
                    enter_file_name_for_compression: "输入要压缩的文件名",
                    toggle_fullscreen: "切换全屏",
                    edit_file: "编辑文件",
                    file_content: "文件内容",
                    loading: "加载中",
                    search: "搜索",
                    create_folder: "创建文件夹",
                    create: "创建",
                    folder_name: "文件夹名称",
                    upload: "上传",
                    change_permissions: "修改权限",
                    change: "修改",
                    details: "详细信息",
                    icons: "图标",
                    list: "列表",
                    name: "名称",
                    size: "尺寸",
                    actions: "操作",
                    date: "日期",
                    selection: "选择",
                    no_files_in_folder: "此文件夹没有文件",
                    no_folders_in_folder: "此文件夹不包含子文件夹",
                    select_this: "选择此文件",
                    go_back: "后退",
                    wait: "等待",
                    move: "移动",
                    download: "下载",
                    view_item: "查看子项",
                    remove: "删除",
                    edit: "编辑",
                    copy: "复制",
                    rename: "重命名",
                    extract: "解压",
                    compress: "压缩",
                    error_invalid_filename: "非法文件名或文件已经存在, 请指定其它名称",
                    error_modifying: "修改文件出错",
                    error_deleting: "删除文件或文件夹出错",
                    error_renaming: "重命名文件出错",
                    error_copying: "复制文件出错",
                    error_compressing: "压缩文件或文件夹出错",
                    error_extracting: "解压文件出错",
                    error_creating_folder: "创建文件夹出错",
                    error_getting_content: "获取文件内容出错",
                    error_changing_perms: "修改文件权限出错",
                    error_uploading_files: "上传文件出错",
                    sure_to_start_compression_with: "确定要压缩？",
                    owner: "拥有者",
                    group: "群组",
                    others: "其他",
                    read: "读取",
                    write: "写入",
                    exec: "执行",
                    original: "原始",
                    changes: "变化",
                    recursive: "递归",
                    preview: "成员预览",
                    open: "打开",
                    these_elements: "共 {{total}} 个",
                    new_folder: "新文件夹",
                    download_as_zip: "下载的ZIP"
                })
               
        }
    ])
}(angular),
function(e, r) {
    "use strict";
    e.module("FileManagerApp").service("apiHandler", ["$http", "$q", "$window", "$translate", 
        function(e, n, i, a, t) {
            e.defaults.headers.common["X-Requested-With"] = "XMLHttpRequest";
            var o = function() {
                this.inprocess = !1,
                    this.asyncSuccess = !1,
                    this.error = ""
            };
            return o.prototype.deferredHandler = function(e, r, n, i) {
                    return e && "object" == typeof e || (this.error = "Error %s - 请求失败，登录可能已过期，请重新登陆.".replace("%s", n)),
                        404 == n && (this.error = "Error 404 - Backend bridge is not working, please check the ajax response."),
                        e.result && e.result.error && (this.error = e.result.error), !this.error && e.error && (this.error = e.error.message), !this.error && i && (this.error = i),
                        this.error ? r.reject(e) : r.resolve(e)
                },
                o.prototype.list = function(r, i, a, t) {
                    var o = this,
                        s = a || o.deferredHandler,
                        l = n.defer(),
                        d = {
                            action: "list",
                            path: i,
                            fileExtensions: t && t.length ? t : void 0
                        };
                    return o.inprocess = !0,
                        o.error = "",
                        e.post(r, d).success(function(e, r) {
                            s(e, l, r)
                        }).error(function(e, r) {
                            s(e, l, r, "请求失败，登录可能已过期，请重新登录")
                        })["finally"](function() {
                            o.inprocess = !1
                        }),
                        l.promise
                },
                o.prototype.copy = function(r, i, t, o) {
                    var s = this,
                        l = n.defer(),
                        d = {
                            action: "copy",
                            items: i,
                            newPath: t
                        };
                    return o && 1 === i.length && (d.singleFilename = o),
                        s.inprocess = !0,
                        s.error = "",
                        e.post(r, d).success(function(e, r) {
                            s.deferredHandler(e, l, r)
                        }).error(function(e, r) {
                            s.deferredHandler(e, l, r, a.instant("error_copying"))
                        })["finally"](function() {
                            s.inprocess = !1
                        }),
                        l.promise
                },
                o.prototype.move = function(r,dir, i, t) {
                    var o = this,
                        s = n.defer(),
                        l = {
                            action: "move",
                            items: i,
                            dirs:dir,
                            newPath: t
                        };
                    return o.inprocess = !0,
                        o.error = "",
                        e.post(r, l).success(function(e, r) {
                            o.deferredHandler(e, s, r)
                        }).error(function(e, r) {
                            o.deferredHandler(e, s, r, a.instant("error_moving"))
                        })["finally"](function() {
                            o.inprocess = !1
                        }),
                        s.promise
                },
                o.prototype.remove = function(r, i,dir) {
                    var t = this,
                        o = n.defer(),
                        s = {
                            action: "remove",
                            items: i,
                            dirs:dir
                        };
                    return t.inprocess = !0,
                        t.error = "",
                        e.post(r, s).success(function(e, r) {
                            t.deferredHandler(e, o, r)
                        }).error(function(e, r) {
                            t.deferredHandler(e, o, r, a.instant("error_deleting"))
                        })["finally"](function() {
                            t.inprocess = !1
                        }),
                        o.promise
                },

                o.prototype.getContent = function(r, i) {
                    var t = this,
                        o = n.defer(),
                        s = {
                            action: "getContent",
                            item: i
                        };
                    return t.inprocess = !0,
                        t.error = "",
                        e.post(r, s).success(function(e, r) {
                            t.deferredHandler(e, o, r)
                        }).error(function(e, r) {
                            t.deferredHandler(e, o, r, a.instant("error_getting_content"))
                        })["finally"](function() {
                            t.inprocess = !1
                        }),
                        o.promise
                },
                o.prototype.edit = function(r, i, t) {
                    var o = this,
                        s = n.defer(),
                        l = {
                            action: "edit",
                            item: i,
                            content: t
                        };
                    return o.inprocess = !0,
                        o.error = "",
                        e.post(r, l).success(function(e, r) {
                            o.deferredHandler(e, s, r)
                        }).error(function(e, r) {
                            o.deferredHandler(e, s, r, a.instant("error_modifying"))
                        })["finally"](function() {
                            o.inprocess = !1
                        }),
                        s.promise
                },
                o.prototype.rename = function(r, i, t) {
                    var o = this,
                        s = n.defer(),
                        l = {
                            action: "rename",
                            item: i,
                            newItemPath: t
                        };
                    return o.inprocess = !0,
                        o.error = "",
                        e.post(r, l).success(function(e, r) {
                            o.deferredHandler(e, s, r)
                        }).error(function(e, r) {
                            o.deferredHandler(e, s, r, a.instant("error_renaming"))
                        })["finally"](function() {
                            o.inprocess = !1
                        }),
                        s.promise
                },
                o.prototype.sharep = function(r, i) {
                    var o = this,
                        s = n.defer(),
                        l = {
                            action: "share",
                            item: i,
                            shareType: "public"
                        };
                    return o.inprocess = !0,
                        o.error = "",
                        e.post(r, l).success(function(e, r) {
                            o.deferredHandler(e, s, r)
                        }).error(function(e, r) {
                            o.deferredHandler(e, s, r, a.instant("error_renaming"))
                        })["finally"](function() {
                            o.inprocess = !1
                        }),
                        s.promise
                },
                o.prototype.sharec = function(r, i) {
                    var o = this,
                        s = n.defer(),
                        l = {
                            action: "share",
                            item: i,
                            shareType: "private"
                        };
                    return o.inprocess = !0,
                        o.error = "",
                        e.post(r, l).success(function(e, r) {
                            o.deferredHandler(e, s, r)
                        }).error(function(e, r) {
                            o.deferredHandler(e, s, r, a.instant("error_renaming"))
                        })["finally"](function() {
                            o.inprocess = !1
                        }),
                        s.promise
                },
                o.prototype.getUrl = function(e, n) {
                    var i = {
                        action: "download",
                        path: n
                    };
                    return n && [e, r.param(i)].join("?")
                },
                   o.prototype.preview = function(e, n) {
                    var i = {
                        action: "preview",
                        path: n
                    };
                    return n && [e, r.param(i)].join("?")
                },
                o.prototype.listPic = function(e, n) {
                    return "ds";
                },
                o.prototype.getsource = function(e, n) {
                    var i = {
                        action: "source",
                        path: n
                    };
                      $.post(e,i,function(data){
                        var data = eval("("+data+")");
                        document.getElementById("source-target").value=data.url;
                        return n && [e, r.param(i)].join("?")
                      })

                },
                o.prototype.download = function(r, t, o, s, l) {
                    var d = this,
                        c = this.getUrl(r, t);
                    if (!s || l || !i.saveAs) return !i.saveAs && i.console.log("Your browser dont support ajax download, downloading by default"), !!i.open(c, "_blank", "");
                    var p = n.defer();
                    return d.inprocess = !0,
                        e.get(c).success(function(e) {
                            var r = new i.Blob([e]);
                            p.resolve(e),
                                i.saveAs(r, o)
                        }).error(function(e, r) {
                            d.deferredHandler(e, p, r, a.instant("error_downloading"))
                        })["finally"](function() {
                            d.inprocess = !1
                        }),
                        p.promise
                },
                o.prototype.downloadMultiple = function(t, o, s, l, d) {
                    var c = this,
                        p = n.defer(),
                        m = {
                            action: "downloadMultiple",
                            items: o,
                            toFilename: s
                        },
                        u = [t, r.param(m)].join("?");
                    return l && !d && i.saveAs ? (c.inprocess = !0, e.get(t).success(function(e) {
                        var r = new i.Blob([e]);
                        p.resolve(e),
                            i.saveAs(r, s)
                    }).error(function(e, r) {
                        c.deferredHandler(e, p, r, a.instant("error_downloading"))
                    })["finally"](function() {
                        c.inprocess = !1
                    }), p.promise) : (!i.saveAs && i.console.log("Your browser dont support ajax download, downloading by default"), !!i.open(u, "_blank", ""))
                },
                o.prototype.compress = function(r, i, t, o) {
                    var s = this,
                        l = n.defer(),
                        d = {
                            action: "compress",
                            items: i,
                            destination: o,
                            compressedFilename: t
                        };
                    return s.inprocess = !0,
                        s.error = "",
                        e.post(r, d).success(function(e, r) {
                            s.deferredHandler(e, l, r)
                        }).error(function(e, r) {
                            s.deferredHandler(e, l, r, a.instant("error_compressing"))
                        })["finally"](function() {
                            s.inprocess = !1
                        }),
                        l.promise
                },
                o.prototype.extract = function(r, i, t, o) {
                    var s = this,
                        l = n.defer(),
                        d = {
                            action: "extract",
                            item: i,
                            destination: o,
                            folderName: t
                        };
                    return s.inprocess = !0,
                        s.error = "",
                        e.post(r, d).success(function(e, r) {
                            s.deferredHandler(e, l, r)
                        }).error(function(e, r) {
                            s.deferredHandler(e, l, r, a.instant("error_extracting"))
                        })["finally"](function() {
                            s.inprocess = !1
                        }),
                        l.promise
                },
                o.prototype.changePermissions = function(r, i, t, o, s) {
                    var l = this,
                        d = n.defer(),
                        c = {
                            action: "changePermissions",
                            items: i,
                            perms: t,
                            permsCode: o,
                            recursive: !!s
                        };
                    return l.inprocess = !0,
                        l.error = "",
                        e.post(r, c).success(function(e, r) {
                            l.deferredHandler(e, d, r)
                        }).error(function(e, r) {
                            l.deferredHandler(e, d, r, a.instant("error_changing_perms"))
                        })["finally"](function() {
                            l.inprocess = !1
                        }),
                        d.promise
                },
                o.prototype.createFolder = function(r, i) {
                    var t = this,
                        o = n.defer(),
                        s = {
                            action: "createFolder",
                            newPath: i
                        };
                    return t.inprocess = !0,
                        t.error = "",
                        e.post(r, s).success(function(e, r) {
                            t.deferredHandler(e, o, r)
                        }).error(function(e, r) {
                            t.deferredHandler(e, o, r, a.instant("error_creating_folder"))
                        })["finally"](function() {
                            t.inprocess = !1
                        }),
                        o.promise
                },
                o
        }
    ])
}(angular, jQuery),
function(e) {
    "use strict";
    e.module("FileManagerApp").service("apiMiddleware", ["$window", "fileManagerConfig", "apiHandler",
        function(e, r, n) {
            var i = function() {
                this.apiHandler = new n
            };
            return i.prototype.getPath = function(e) {
                    return "/" + e.join("/")
                },
                i.prototype.getFileList = function(e) {
                    return (e || []).map(function(e) {
                        return e && e.model.fullPath()
                    })
                },
                i.prototype.getFilePath = function(e) {
                    return e && e.model.fullPath()
                },
                i.prototype.list = function(e, n) {
                    return this.apiHandler.list(r.listUrl, this.getPath(e), n)
                },
                i.prototype.copy = function(e, n) {
                    var i = this.getFileList(e),
                        a = 1 === i.length ? e[0].tempModel.name : void 0;
                    return this.apiHandler.copy(r.copyUrl, i, this.getPath(n), a)
                },
                i.prototype.move = function(e,dir, n) {
                    var i = this.getFileList(e);
                    var dirList = this.getFileList(dir);
                    return this.apiHandler.move(r.moveUrl,dirList ,i, this.getPath(n))
                },
                i.prototype.remove = function(e,dir) {
                    var n = this.getFileList(e);
                    var dirList = this.getFileList(dir);
                    return this.apiHandler.remove(r.removeUrl, n,dirList)
                },
              
                i.prototype.getContent = function(e) {
                    var n = this.getFilePath(e);
                    return this.apiHandler.getContent(r.getContentUrl, n)
                },
                i.prototype.edit = function(e) {
                    var n = this.getFilePath(e);
                    return this.apiHandler.edit(r.editUrl, n, e.tempModel.content)
                },
                i.prototype.rename = function(e) {
                    var n = this.getFilePath(e),
                        i = e.tempModel.fullPath();
                    return this.apiHandler.rename(r.renameUrl, n, i)
                },
                i.prototype.sharep = function(e) {
                    var n = this.getFilePath(e);
                    return this.apiHandler.sharep(r.shareUrl, n)
                },
                i.prototype.sharec = function(e) {
                    var n = this.getFilePath(e);
                    return this.apiHandler.sharec(r.shareUrl, n)
                },
                i.prototype.getUrl = function(e) {
                    var n = this.getFilePath(e);
                    return this.apiHandler.getUrl(r.downloadFileUrl, n)
                },
                 i.prototype.preview = function(e) {
                    var n = this.getFilePath(e);
                    return this.apiHandler.preview(r.previewUrl, n)
                },
                i.prototype.listPic = function(e) {
                    var n = this.getFilePath(e);
                     var s = true;
                    $.get({async:false,url:"/Share/ListPic?path="+n+"&id="+shareInfo.shareId}).complete(function(data){
　　                  s =  data;
                    });
                    return s.responseJSON;
            },
                
                i.prototype.getsource = function(e) {
                    var n = this.getFilePath(e);
                    return this.apiHandler.getsource(r.sourceUrl, n)
                },
                i.prototype.download = function(e, n) {
                    var i = this.getFilePath(e),
                        a = e.model.name;
                    return e.isFolder() ? void 0 : this.apiHandler.download(r.downloadFileUrl, i, a, r.downloadFilesByAjax, n)
                },
                i.prototype.downloadMultiple = function(e, n) {
                    var i = this.getFileList(e),
                        a = (new Date).getTime().toString().substr(8, 13),
                        t = a + "-" + r.multipleDownloadFileName;
                    return this.apiHandler.downloadMultiple(r.downloadMultipleUrl, i, t, r.downloadFilesByAjax, n)
                },
                i.prototype.compress = function(e, n, i) {
                    var a = this.getFileList(e);
                    return this.apiHandler.compress(r.compressUrl, a, n, this.getPath(i))
                },
                i.prototype.extract = function(e, n, i) {
                    var a = this.getFilePath(e);
                    return this.apiHandler.extract(r.extractUrl, a, n, this.getPath(i))
                },
                i.prototype.changePermissions = function(e, n) {
                    var i = this.getFileList(e),
                        a = n.tempModel.perms.toCode(),
                        t = n.tempModel.perms.toOctal(),
                        o = !!n.tempModel.recursive;
                    return this.apiHandler.changePermissions(r.permissionsUrl, i, a, t, o)
                },
                i.prototype.createFolder = function(e) {
                    var n = e.tempModel.fullPath();
                    return this.apiHandler.createFolder(r.createFolderUrl, n)
                },
                i
        }
    ])
}(angular),
function(e) {
    "use strict";
    e.module("FileManagerApp").service("fileNavigator", ["apiMiddleware", "fileManagerConfig", "item",
        function(e, r, n) {
            var i = function() {
                this.apiMiddleware = new e,
                    this.requesting = !1,
                    this.fileList = [],
                    this.currentPath = this.getBasePath(),
                    this.history = [],
                    this.error = "",
                    this.onRefresh = function() {}
            };
            return i.prototype.getBasePath = function() {
                    var e = (r.basePath || "").replace(/^\//, "");
                    return e.trim() ? e.split("/") : []
                },
                i.prototype.deferredHandler = function(e, r, n, i) {
                    return e && "object" == typeof e || (this.error = "Error %s - 请求失败，登录可能已过期，请重新登录".replace("%s", n)),
                        404 == n && (this.error = "Error 404 - Backend bridge is not working, please check the ajax response."),
                        200 == n && (this.error = null), !this.error && e.result && e.result.error && (this.error = e.result.error), !this.error && e.error && (this.error = e.error.message), !this.error && i && (this.error = i),
                        this.error ? r.reject(e) : r.resolve(e)
                },
                i.prototype.list = function() {
                    return this.apiMiddleware.list(this.currentPath, this.deferredHandler.bind(this))
                },
                i.prototype.refresh = function() {
                    var e = this;
                    e.currentPath.length || (e.currentPath = this.getBasePath());
                    var r = e.currentPath.join("/");
                    return e.requesting = !0,
                        e.fileList = [],
                        e.list().then(function(i) {
                            e.fileList = (i.result || []).map(function(r) {
                                    return new n(r, e.currentPath)
                                }),
                                e.buildTree(r),
                                e.onRefresh()
                        })["finally"](function() {
                            e.requesting = !1
                        })
                },
                i.prototype.buildTree = function(e) {
                    function r(e, n, i) {
                        var a = i ? i + "/" + n.model.name : n.model.name;
                        if (e.name && e.name.trim() && 0 !== i.trim().indexOf(e.name) && (e.nodes = []), e.name !== i) e.nodes.forEach(function(e) {
                            r(e, n, i)
                        });
                        else {
                            for (var t in e.nodes)
                                if (e.nodes[t].name === a) return;
                            e.nodes.push({
                                item: n,
                                name: a,
                                nodes: []
                            })
                        }
                        e.nodes = e.nodes.sort(function(e, r) {
                            return e.name.toLowerCase() < r.name.toLowerCase() ? -1 : e.name.toLowerCase() === r.name.toLowerCase() ? 0 : 1
                        })
                    }

                    function i(e, r) {
                        r.push(e);
                        for (var n in e.nodes) i(e.nodes[n], r)
                    }

                    function a(e, r) {
                        return e.filter(function(e) {
                            return e.name === r
                        })[0]
                    }
                    var t = [],
                        o = {};
                    !this.history.length && this.history.push({
                            name: this.getBasePath()[0] || "",
                            nodes: []
                        }),
                        i(this.history[0], t),
                        o = a(t, e),
                        o && (o.nodes = []);
                    for (var s in this.fileList) {
                        var l = this.fileList[s];
                        l instanceof n && l.isFolder() && r(this.history[0], l, e)
                    }
                },
                i.prototype.folderClick = function(e) {
                    this.currentPath = [],
                        e && e.isFolder() && (this.currentPath = e.model.fullPath().split("/").splice(1)),
                        this.refresh()
                },
                i.prototype.upDir = function() {
                    this.currentPath[0] && (this.currentPath = this.currentPath.slice(0, -1), this.refresh())
                },
                i.prototype.goTo = function(e) {
                    this.currentPath = this.currentPath.slice(0, e + 1),
                        this.refresh()
                },
                i.prototype.fileNameExists = function(e) {
                    return this.fileList.find(function(r) {
                        return e && r.model.name.trim() === e.trim()
                    })
                },
                i.prototype.listHasFolders = function() {
                    return this.fileList.find(function(e) {
                        return "dir" === e.model.type
                    })
                },
                i.prototype.getCurrentFolderName = function() {
                    return this.currentPath.slice(-1)[0] || "/"
                },
                i
        }
    ])
}(angular),
angular.module("FileManagerApp").run(["$templateCache",
    function(e) {
        e.put("src/templates/current-folder-breadcrumb.html", '<ol class="breadcrumb">\r\n    <li>\r\n        <a href="" class="wave_hide" ng-click="fileNavigator.goTo(-1)">\r\n            {{"filemanager" | translate}}\r\n        </a>\r\n    </li>\r\n    <li ng-repeat="(key, dir) in fileNavigator.currentPath track by key" ng-class="{\'active\':$last}" class="animated fast fadeIn">\r\n        <a href="" ng-show="!$last" ng-click="fileNavigator.goTo(key)" class="notWave">\r\n            {{dir | strLimit : 8}}\r\n        </a>\r\n        <span ng-show="$last">\r\n            {{dir | strLimit : 12}}\r\n        </span>\r\n    </li>\r\n</ol>'),
            e.put("src/templates/item-context-menu.html", '<div id="context-menu" class="dropdown clearfix animated fast fadeIn">\r\n    <ul class="dropdown-menu dropdown-right-click" role="menu" aria-labelledby="dropdownMenu" ng-show="temps.length">\r\n\r\n        <li ng-show="singleSelection() && singleSelection().isFolder()">\r\n            <a href="" tabindex="-1" ng-click="smartClick(singleSelection())">\r\n                <i class="glyphicon glyphicon-folder-open"></i> {{\'open\' | translate}}\r\n            </a>\r\n        </li>\r\n\r\n        <li ng-show="config.pickCallback && singleSelection() && singleSelection().isSelectable()">\r\n            <a href="" tabindex="-1" ng-click="config.pickCallback(singleSelection().model)">\r\n                <i class="glyphicon glyphicon-hand-up"></i> {{\'select_this\' | translate}}\r\n            </a>\r\n        </li>\r\n\r\n        <li ng-show="config.allowedActions.download && !selectionHas(\'dir\') && singleSelection()">\r\n            <a href="" tabindex="-1" ng-click="download()">\r\n             <i class="glyphicon glyphicon-cloud-download"></i> {{\'download\' | translate}}\r\n            </a>\r\n        </li>\r\n\r\n   <li ng-show="config.allowedActions.getSource && !selectionHas(\'dir\') && singleSelection()">\r\n            <a href="" tabindex="-1" ng-click="openGetSource()">\r\n             <i class="glyphicon glyphicon-link"></i> 获取外链\r\n            </a>\r\n        </li> <li ng-show="config.allowedActions.shareFile && singleSelection()">\r\n            <a href="" tabindex="-1" ng-click="modal(\'share\')">\r\n             <i class="glyphicon glyphicon-share"></i> 分享\r\n            </a>\r\n        </li><li ng-show="config.allowedActions.downloadMultiple && !selectionHas(\'dir\') && !singleSelection()">\r\n            <a href="" tabindex="-1" ng-click="download()">\r\n                <i class="glyphicon glyphicon-cloud-download"></i> {{\'download_as_zip\' | translate}}\r\n            </a>\r\n        </li>\r\n\r\n        <li ng-show="config.allowedActions.preview && singleSelection().isImage() && singleSelection()">\r\n            <a href="" tabindex="-1" ng-click="openImagePreview()">\r\n                <i class="glyphicon glyphicon-picture"></i> 预览图像\r\n            </a>\r\n        </li>\r\n\r\n     <li ng-show="config.allowedActions.preview && singleSelection().isVideo() && singleSelection()">\r\n            <a href="" tabindex="-1" ng-click="openVideoPreview()">\r\n                <i class="glyphicon glyphicon-facetime-video"></i> 预览视频\r\n            </a>\r\n        </li>  <li ng-show="config.allowedActions.preview && singleSelection().isAudio() && singleSelection()">\r\n            <a href="" tabindex="-1" ng-click="openAudioPreview()">\r\n                <i class="glyphicon glyphicon-music"></i> 预览音频\r\n            </a>\r\n        </li>  <li ng-show="config.allowedActions.rename && singleSelection()">\r\n            <a href="" tabindex="-1" ng-click="modal(\'rename\')">\r\n                <i class="glyphicon glyphicon-edit"></i> {{\'rename\' | translate}}\r\n            </a>\r\n        </li>\r\n\r\n        <li ng-show="config.allowedActions.move">\r\n            <a href="" tabindex="-1" ng-click="modalWithPathSelector(\'move\')">\r\n                <i class="glyphicon glyphicon-arrow-right"></i> {{\'move\' | translate}}\r\n            </a>\r\n        </li>\r\n\r\n        <li ng-show="config.allowedActions.copy && !selectionHas(\'dir\')">\r\n            <a href="" tabindex="-1" ng-click="modalWithPathSelector(\'copy\')">\r\n                <i class="glyphicon glyphicon-log-out"></i> {{\'copy\' | translate}}\r\n            </a>\r\n        </li>\r\n\r\n        <li ng-show="config.allowedActions.edit && singleSelection() && singleSelection().isEditable()">\r\n            <a href="" tabindex="-1" ng-click="openEditItem()">\r\n                <i class="glyphicon glyphicon-pencil"></i> {{\'edit\' | translate}}\r\n            </a>\r\n        </li>\r\n\r\n        <li ng-show="config.allowedActions.changePermissions">\r\n            <a href="" tabindex="-1" ng-click="modal(\'changepermissions\')">\r\n                <i class="glyphicon glyphicon-lock"></i> {{\'permissions\' | translate}}\r\n            </a>\r\n        </li>\r\n\r\n        <li ng-show="config.allowedActions.compress && (!singleSelection() || selectionHas(\'dir\'))">\r\n            <a href="" tabindex="-1" ng-click="modal(\'compress\')">\r\n                <i class="glyphicon glyphicon-compressed"></i> {{\'compress\' | translate}}\r\n            </a>\r\n        </li>\r\n\r\n        <li ng-show="config.allowedActions.extract && singleSelection() && singleSelection().isExtractable()">\r\n            <a href="" tabindex="-1" ng-click="modal(\'extract\')">\r\n                <i class="glyphicon glyphicon-export"></i> {{\'extract\' | translate}}\r\n            </a>\r\n        </li>\r\n\r\n        <li class="divider" ng-show="config.allowedActions.remove"></li>\r\n        \r\n        <li ng-show="config.allowedActions.remove">\r\n            <a href="" tabindex="-1" ng-click="modal(\'remove\')">\r\n                <i class="glyphicon glyphicon-trash"></i> {{\'remove\' | translate}}\r\n            </a>\r\n        </li>\r\n\r\n    </ul>\r\n\r\n    \r\n</div>'),
            e.put("src/templates/main-icons.html", '<div class="iconset noselect">\r\n    <div class="item-list clearfix" ng-click="selectOrUnselect(null, $event)" ng-right-click="selectOrUnselect(null, $event)" prevent="true">\r\n        <div class="col-120" ng-repeat="item in $parent.fileList = (fileNavigator.fileList | filter: {model:{name: query}})" ng-show="!fileNavigator.requesting && !fileNavigator.error">\r\n            <a href="" class="thumbnail text-center withripple" ng-click="selectOrUnselect(item, $event)" ng-dblclick="smartClick(item)" ng-right-click="selectOrUnselect(item, $event)" title="{{item.model.name}} ({{item.model.size | humanReadableFileSize}})" ng-class="{selected: isSelected(item)}">\r\n                <div class="item-icon">\r\n                    <i class="glyphicon glyphicon-folder-open" ng-show="item.model.type === \'dir\'"></i>\r\n                    <i class="glyphicon glyphicon-facetime-video icon-black" data-ext="{{ item.model.name | fileExtension }}" ng-show="item.isVideo()&&item.model.type === \'file\'" ng-class=""></i>\r\n   <img class="smallImg" src="/Share/Thumb/?path={{ item.model.encodedPath()}}&isImg={{item.isImage()}}&shareKey={{item.getShareKey()}}" data-ext="{{ item.model.name | fileExtension }}" ng-show="item.model.pic!=\'\'&&item.isImage()&&item.model.type === \'file\'" ng-class=""></img>\r\n   <i class="glyphicon glyphicon-music icon-black" data-ext="{{ item.model.name | fileExtension }}" ng-show="item.isAudio()&&item.model.type === \'file\'" ng-class=""></i>\r\n <i class="glyphicon glyphicon-picture icon-black" data-ext="{{ item.model.name | fileExtension }}" ng-show="item.model.pic==\'\'&&item.isImage()&&item.model.type === \'file\'" ng-class=""></i>\r\n      <i class="glyphicon glyphicon-file icon-black" data-ext="{{ item.model.name | fileExtension }}" ng-show="!item.isAudio()&&!item.isImage()&&!item.isVideo()&&item.model.type === \'file\'" ng-class="{\'item-extension\': config.showExtensionIcons}"></i>\r\n      </div>\r\n                <span ng-show="item.model.type != \'dir\'" class="file_name icon-black">{{item.model.name | strLimit : 11 }}</span>\r\n   <span ng-show="item.model.type === \'dir\'" class="file_name ">{{item.model.name | strLimit : 11 }}</span>\r\n          </a>\r\n        </div>\r\n    </div>\r\n\r\n    <div ng-show="fileNavigator.requesting">\r\n        <div ng-include="config.tplPath + \'/spinner.html\'"></div>\r\n    </div>\r\n\r\n    <div class="alert alert-warning" ng-show="!fileNavigator.requesting && fileNavigator.fileList.length < 1 && !fileNavigator.error">\r\n        {{"no_files_in_folder" | translate}}...\r\n    </div>\r\n    \r\n    <div class="alert alert-danger" ng-show="!fileNavigator.requesting && fileNavigator.error">\r\n        {{ fileNavigator.error }}\r\n    </div>\r\n</div>'),
            e.put("src/templates/main-table-modal.html", '<table class="table table-condensed table-modal-condensed mb0">\r\n    <thead>\r\n        <tr>\r\n            <th>\r\n                <a href="" ng-click="order(\'model.name\')">\r\n                    {{"name" | translate}}\r\n                    <span class="sortorder" ng-show="predicate[1] === \'model.name\'" ng-class="{reverse:reverse}"></span>\r\n                </a>\r\n            </th>\r\n            <th class="text-right"></th>\r\n        </tr>\r\n    </thead>\r\n    <tbody class="file-item">\r\n        <tr ng-show="fileNavigator.requesting">\r\n            <td colspan="2">\r\n                <div ng-include="config.tplPath + \'/spinner.html\'"></div>\r\n            </td>\r\n        </tr>\r\n        <tr ng-show="!fileNavigator.requesting && !fileNavigator.listHasFolders() && !fileNavigator.error">\r\n            <td>\r\n                {{"no_folders_in_folder" | translate}}...\r\n            </td>\r\n            <td class="text-right">\r\n                <button class="btn btn-sm btn-default" ng-click="fileNavigator.upDir()">{{"go_back" | translate}}</button>\r\n            </td>\r\n        </tr>\r\n        <tr ng-show="!fileNavigator.requesting && fileNavigator.error">\r\n            <td colspan="2">\r\n                {{ fileNavigator.error }}\r\n            </td>\r\n        </tr>\r\n        <tr ng-repeat="item in fileNavigator.fileList | orderBy:predicate:reverse" ng-show="!fileNavigator.requesting && item.model.type === \'dir\'" ng-if="!selectedFilesAreChildOfPath(item)">\r\n            <td>\r\n                <a href="" ng-click="fileNavigator.folderClick(item)" title="{{item.model.name}} ({{item.model.size | humanReadableFileSize}})">\r\n                    <i class="glyphicon glyphicon-folder-close"></i>\r\n                    {{item.model.name | strLimit : 32}}\r\n                </a>\r\n            </td>\r\n            <td class="text-right">\r\n                <button class="btn btn-sm btn-default" ng-click="select(item)">\r\n                    <i class="glyphicon glyphicon-hand-up"></i> {{"select_this" | translate}}\r\n                </button>\r\n            </td>\r\n        </tr>\r\n    </tbody>\r\n</table>'),
            e.put("src/templates/main-table.html", '<table class="table mb0 table-files noselect">\r\n    <thead>\r\n        <tr>\r\n            <th>\r\n                <a href="" ng-click="order(\'model.name\')">\r\n                    {{"name" | translate}}\r\n                    <span class="sortorder" ng-show="predicate[1] === \'model.name\'" ng-class="{reverse:reverse}"></span>\r\n                </a>\r\n            </th>\r\n            <th class="hidden-xs" ng-hide="config.hideSize">\r\n                <a href="" ng-click="order(\'model.size\')">\r\n                    {{"size" | translate}}\r\n                    <span class="sortorder" ng-show="predicate[1] === \'model.size\'" ng-class="{reverse:reverse}"></span>\r\n                </a>\r\n            </th>\r\n            <th class="hidden-sm hidden-xs" ng-hide="config.hideDate">\r\n                <a href="" ng-click="order(\'model.date\')">\r\n                    {{"date" | translate}}\r\n                    <span class="sortorder" ng-show="predicate[1] === \'model.date\'" ng-class="{reverse:reverse}"></span>\r\n                </a>\r\n            </th>\r\n            <th class="hidden-sm hidden-xs" ng-hide="config.hidePermissions">\r\n                <a href="" ng-click="order(\'model.permissions\')">\r\n                    父目录\r\n                    <span class="sortorder" ng-show="predicate[1] === \'model.permissions\'" ng-class="{reverse:reverse}"></span>\r\n                </a>\r\n            </th>\r\n        </tr>\r\n    </thead>\r\n    <tbody class="file-item">\r\n        <tr ng-show="fileNavigator.requesting">\r\n            <td colspan="5">\r\n                <div ng-include="config.tplPath + \'/spinner.html\'"></div>\r\n            </td>\r\n        </tr>\r\n        <tr ng-show="!fileNavigator.requesting &amp;&amp; fileNavigator.fileList.length < 1 &amp;&amp; !fileNavigator.error">\r\n            <td colspan="5">\r\n                {{"no_files_in_folder" | translate}}...\r\n            </td>\r\n        </tr>\r\n        <tr ng-show="!fileNavigator.requesting &amp;&amp; fileNavigator.error">\r\n            <td colspan="5">\r\n                {{ fileNavigator.error }}\r\n            </td>\r\n        </tr>\r\n        <tr class="item-list" ng-repeat="item in $parent.fileList = (fileNavigator.fileList | filter: {model:{name: query}} | orderBy:predicate:reverse)" ng-show="!fileNavigator.requesting" ng-click="selectOrUnselect(item, $event)" ng-dblclick="smartClick(item)" ng-right-click="selectOrUnselect(item, $event)" ng-class="{selected: isSelected(item)}">\r\n            <td>\r\n                <a href="" title="{{item.model.name}} ({{item.model.size | humanReadableFileSize}})">\r\n                    <i class="glyphicon glyphicon-folder-close" ng-show="item.model.type === \'dir\'"></i>\r\n                    <i class="glyphicon glyphicon-file" ng-show="item.model.type === \'file\'"></i>\r\n                    {{item.model.name | strLimit : 64}}\r\n                </a>\r\n            </td>\r\n            <td class="hidden-xs">\r\n                <span ng-show="item.model.type !== \'dir\' || config.showSizeForDirectories">\r\n                    {{item.model.size | humanReadableFileSize}}\r\n                </span>\r\n            </td>\r\n            <td class="hidden-sm hidden-xs" ng-hide="config.hideDate">\r\n                {{item.model.date | formatDate }}\r\n            </td>\r\n            <td class="hidden-sm hidden-xs" ng-hide="config.hidePermissions">\r\n              {{item.model.name2 | strLimit : 64}}\r\n            </td>\r\n        </tr>\r\n    </tbody>\r\n</table>\r\n'),
            e.put("src/templates/main.html", '<div ng-controller="FileManagerCtrl" class="file-main">\r\n    <div ng-include="config.tplPath + \'/navbar.html\'"></div>\r\n\r\n    <div class="container-fluid">\r\n        <div class="row">\r\n\r\n            <div class="col-sm-4 col-md-2 sidebar file-tree animated slow fadeIn lefts" ng-include="config.tplPath + \'/sidebar.html\'" ng-show="config.sidebar &amp;&amp; fileNavigator.history[0]">\r\n            </div>\r\n\r\n            <div class="main" ng-class="config.sidebar &amp;&amp; fileNavigator.history[0] &amp;&amp; \'col-sm-8 col-md-10\'">\r\n                <div ng-include="config.tplPath + \'/\' + viewTemplate" class="main-navigation clearfix"></div>\r\n            </div>\r\n        </div>\r\n    </div>\r\n\r\n    <div ng-include="config.tplPath + \'/modals.html\'"></div>\r\n    <div ng-include="config.tplPath + \'/item-context-menu.html\'"></div>\r\n</div>\r\n'),
            e.put("src/templates/modals.html", '<div class="modal animated fadeIn" id="imagepreview">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n      <div class="modal-header">\r\n        <button type="button" class="close" data-dismiss="modal">\r\n            <span aria-hidden="true">&times;</span>\r\n            <span class="sr-only">{{"close" | translate}}</span>\r\n        </button>\r\n        <h4 class="modal-title">{{"preview" | translate}}</h4>\r\n      </div>\r\n      <div class="modal-body">\r\n        <div class="text-center">\r\n          <img id="imagepreview-target" class="preview" alt="{{singleSelection().model.name}}" ng-class="{\'loading\': apiMiddleware.apiHandler.inprocess}">\r\n          <span class="label label-warning" ng-show="apiMiddleware.apiHandler.inprocess">{{\'loading\' | translate}} ...</span>\r\n        </div>\r\n        <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n      </div>\r\n      <div class="modal-footer">\r\n        <button type="button" class="btn btn-default" data-dismiss="modal" ng-disabled="apiMiddleware.apiHandler.inprocess">{{"close" | translate}}</button>\r\n      </div>\r\n    </div>\r\n  </div>\r\n</div>\r\n\r\n   <div class="modal animated fadeIn" id="getsource">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n      <div class="modal-header">\r\n        <button type="button" class="close" data-dismiss="modal">\r\n            <span aria-hidden="true">&times;</span>\r\n            <span class="sr-only">{{"close" | translate}}</span>\r\n        </button>\r\n        <h4 class="modal-title">获取外链</h4>\r\n      </div>\r\n      <div class="modal-body">\r\n     <div class=""><lable> {{singleSelection() && singleSelection().model.name}}  的源文件地址：</lable>   \r\n          <input type="text" id="source-target" spellcheck="false" class="form-control" >\r\n            </div>\r\n        <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n      </div>\r\n      <div class="modal-footer">\r\n  <button type="button" class="btn btn-primary btn-copy"  data-clipboard-target="#source-target">复制URL</button>      <button type="button" class="btn btn-default" data-dismiss="modal" >{{"close" | translate}}</button>\r\n      </div>\r\n    </div>\r\n  </div>\r\n</div>\r\n\r\n            <div class="modal animated fadeIn" id="videopreview">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n      <div class="modal-header">\r\n        <button type="button" class="close" onclick="audioPause()" data-dismiss="modal">\r\n            <span aria-hidden="true">&times;</span>\r\n            <span class="sr-only">{{"close" | translate}}</span>\r\n        </button>\r\n        <h4 class="modal-title">视频预览</h4>\r\n      </div>\r\n      <div class="modal-body">\r\n      <div class="text-center">\r\n          <div  id="videopreview-target" style="width: 100%;object-fit: fill"  class="preview" alt="{{singleSelection().model.name}}" ng-class=""></div>\r\n          <span class="label label-warning" ng-show="apiMiddleware.apiHandler.inprocess">{{\'loading\' | translate}} ...</span>\r\n        </div>\r\n        <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n      </div>\r\n        </div>\r\n  </div>\r\n</div>\r\n\r\n            <div class="modal animated fadeIn" id="audiopreview">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n      <div class="modal-header">\r\n        <button type="button" class="close" data-dismiss="modal" onclick="audioPause()">\r\n            <span aria-hidden="true">&times;</span>\r\n            <span class="sr-only">{{"close" | translate}}</span>\r\n        </button>\r\n        <h4 class="modal-title">音频预览</h4>\r\n      </div>\r\n      <div class="modal-body">\r\n        <div class="text-center">\r\n          <audio  id="audiopreview-target" style="width: 100%;object-fit: fill" controls="controls" class="preview" alt="{{singleSelection().model.name}}" ng-class=""></audio>\r\n          <span class="label label-warning" ng-show="apiMiddleware.apiHandler.inprocess">{{\'loading\' | translate}} ...</span>\r\n        </div>\r\n        <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n      </div>\r\n        </div>\r\n  </div>\r\n</div>\r\n\r\n                <div class="modal animated fadeIn" id="remove">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n    <form ng-submit="remove()">\r\n      <div class="modal-header">\r\n        <button type="button" class="close" data-dismiss="modal">\r\n            <span aria-hidden="true">&times;</span>\r\n            <span class="sr-only">{{"close" | translate}}</span>\r\n        </button>\r\n        <h4 class="modal-title">{{"confirm" | translate}}</h4>\r\n      </div>\r\n      <div class="modal-body">\r\n        {{\'sure_to_delete\' | translate}} <span ng-include data-src="\'selected-files-msg\'"></span>\r\n\r\n        <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n      </div>\r\n      <div class="modal-footer">\r\n        <button type="button" class="btn btn-default" data-dismiss="modal" ng-disabled="apiMiddleware.apiHandler.inprocess">{{"cancel" | translate}}</button>\r\n        <button type="submit" class="btn btn-primary" ng-disabled="apiMiddleware.apiHandler.inprocess" autofocus="autofocus">{{"remove" | translate}}</button>\r\n      </div>\r\n      </form>\r\n    </div>\r\n  </div>\r\n</div>\r\n\r\n<div class="modal animated fadeIn" id="move">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n        <form ng-submit="move()">\r\n            <div class="modal-header">\r\n              <button type="button" class="close" data-dismiss="modal">\r\n                  <span aria-hidden="true">&times;</span>\r\n                  <span class="sr-only">{{"close" | translate}}</span>\r\n              </button>\r\n              <h4 class="modal-title">{{\'move\' | translate}}</h4>\r\n            </div>\r\n            <div class="modal-body">\r\n              <div ng-include data-src="\'path-selector\'" class="clearfix"></div>\r\n              <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n            </div>\r\n            <div class="modal-footer">\r\n              <button type="button" class="btn btn-default" data-dismiss="modal" ng-disabled="apiMiddleware.apiHandler.inprocess">{{"cancel" | translate}}</button>\r\n              <button type="submit" class="btn btn-primary" ng-disabled="apiMiddleware.apiHandler.inprocess">{{\'move\' | translate}}</button>\r\n            </div>\r\n        </form>\r\n    </div>\r\n  </div>\r\n</div>\r\n\r\n\r\n<div class="modal animated fadeIn" id="rename">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n        <form ng-submit="rename()">\r\n            <div class="modal-header">\r\n              <button type="button" class="close" data-dismiss="modal">\r\n                  <span aria-hidden="true">&times;</span>\r\n                  <span class="sr-only">{{"close" | translate}}</span>\r\n              </button>\r\n              <h4 class="modal-title">{{\'rename\' | translate}}</h4>\r\n            </div>\r\n            <div class="modal-body">\r\n              <label class="radio">{{\'enter_new_name_for\' | translate}} <b>{{singleSelection() && singleSelection().model.name}}</b></label>\r\n              <input class="form-control" ng-model="singleSelection().tempModel.name" spellcheck="false" autofocus="autofocus">\r\n\r\n              <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n            </div>\r\n            <div class="modal-footer">\r\n              <button type="button" class="btn btn-default" data-dismiss="modal" ng-disabled="apiMiddleware.apiHandler.inprocess">{{"cancel" | translate}}</button>\r\n              <button type="submit" class="btn btn-primary" ng-disabled="apiMiddleware.apiHandler.inprocess">{{\'rename\' | translate}}</button>\r\n            </div>\r\n        </form>\r\n    </div>\r\n  </div>\r\n</div><div class="modal animated fadeIn" id="share">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n        <form >\r\n            <div class="modal-header">\r\n              <button type="button" class="close" data-dismiss="modal">\r\n                  <span aria-hidden="true">&times;</span>\r\n                  <span class="sr-only">{{"close" | translate}}</span>\r\n              </button>\r\n              <h4 class="modal-title">创建分享</h4>\r\n            </div>\r\n            <div class="modal-body">\r\n              <label class="radio">请选择分享方式：</label>\r\n       <div class="form-group is-empty"> <div class="col-md-6"><button type="button" ng-disabled="apiMiddleware.apiHandler.inprocess" ng-click="sharePublic()" class="btn btn-default" style=" width: 100%; height: 150px;"><i class="glyphicon glyphicon-eye-open" style="font-size: 60px;"></i><br><span>公开分享</span></button></div><div class="col-md-6"><button type="button"  ng-disabled="apiMiddleware.apiHandler.inprocess" ng-click="shareSecret()" class="btn btn-default" style="width: 100%;height: 150px;"><i class="glyphicon glyphicon-eye-close" style="font-size: 60px;"></i><br><span style="">私密分享</span></button></div></div>          <div ng-include data-src="\'error-bar\'" class="clearfix"></div>    <div class="modal-footer">\r\n              <button type="button" class="btn btn-default" data-dismiss="modal" ng-disabled="apiMiddleware.apiHandler.inprocess">{{"cancel" | translate}}</button>\r\n                  </div>\r\n        </form>\r\n    </div>\r\n  </div>\r\n</div></div> <div class="modal animated fadeIn" id="share_result">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n      <div class="modal-header">\r\n        <button type="button" class="close" data-dismiss="modal">\r\n            <span aria-hidden="true">&times;</span>\r\n            <span class="sr-only">{{"close" | translate}}</span>\r\n        </button>\r\n        <h4 class="modal-title">公开分享</h4>\r\n      </div>\r\n      <div class="modal-body">\r\n     <div class=""><lable> <span id="file_name"></span> 的分享地址：</lable>   \r\n          <input type="text" id="share-public-target" spellcheck="false" class="form-control" >\r\n            </div>\r\n        <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n      </div>\r\n      <div class="modal-footer">\r\n  <button type="button" class="btn btn-primary btn-copy"  data-clipboard-target="#share-public-target">复制URL</button>      <button type="button" class="btn btn-default" data-dismiss="modal" >{{"close" | translate}}</button>\r\n      </div>\r\n    </div>\r\n  </div>\r\n</div>\r\n\r\n \r\n\r\n  <div class="modal animated fadeIn" id="copy">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n        <form ng-submit="copy()">\r\n            <div class="modal-header">\r\n              <button type="button" class="close" data-dismiss="modal">\r\n                  <span aria-hidden="true">&times;</span>\r\n                  <span class="sr-only">{{"close" | translate}}</span>\r\n              </button>\r\n              <h4 class="modal-title">{{\'copy_file\' | translate}}</h4>\r\n            </div>\r\n            <div class="modal-body">\r\n              <div ng-show="singleSelection()">\r\n                <label class="radio">{{\'enter_new_name_for\' | translate}} <b>{{singleSelection().model.name}}</b></label>\r\n                <input class="form-control" ng-model="singleSelection().tempModel.name" autofocus="autofocus">\r\n              </div>\r\n\r\n              <div ng-include data-src="\'path-selector\'" class="clearfix"></div>\r\n              <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n            </div>\r\n            <div class="modal-footer">\r\n              <button type="button" class="btn btn-default" data-dismiss="modal" ng-disabled="apiMiddleware.apiHandler.inprocess">{{"cancel" | translate}}</button>\r\n              <button type="submit" class="btn btn-primary" ng-disabled="apiMiddleware.apiHandler.inprocess">{{"copy" | translate}}</button>\r\n            </div>\r\n        </form>\r\n    </div>\r\n  </div>\r\n</div>\r\n\r\n<div class="modal animated fadeIn" id="compress">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n        <form ng-submit="compress()">\r\n            <div class="modal-header">\r\n              <button type="button" class="close" data-dismiss="modal">\r\n                  <span aria-hidden="true">&times;</span>\r\n                  <span class="sr-only">{{"close" | translate}}</span>\r\n              </button>\r\n              <h4 class="modal-title">{{\'compress\' | translate}}</h4>\r\n            </div>\r\n            <div class="modal-body">\r\n              <div ng-show="apiMiddleware.apiHandler.asyncSuccess">\r\n                  <div class="label label-success error-msg">{{\'compression_started\' | translate}}</div>\r\n              </div>\r\n              <div ng-hide="apiMiddleware.apiHandler.asyncSuccess">\r\n                  <div ng-hide="config.allowedActions.compressChooseName">\r\n                    {{\'sure_to_start_compression_with\' | translate}} <b>{{singleSelection().model.name}}</b> ?\r\n                  </div>\r\n                  <div ng-show="config.allowedActions.compressChooseName">\r\n                    <label class="radio">\r\n                      {{\'enter_file_name_for_compression\' | translate}}\r\n                      <span ng-include data-src="\'selected-files-msg\'"></span>\r\n                    </label>\r\n                    <input class="form-control" ng-model="temp.tempModel.name" autofocus="autofocus">\r\n                  </div>\r\n              </div>\r\n\r\n              <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n            </div>\r\n            <div class="modal-footer">\r\n              <div ng-show="apiMiddleware.apiHandler.asyncSuccess">\r\n                  <button type="button" class="btn btn-default" data-dismiss="modal" ng-disabled="apiMiddleware.apiHandler.inprocess">{{"close" | translate}}</button>\r\n              </div>\r\n              <div ng-hide="apiMiddleware.apiHandler.asyncSuccess">\r\n                  <button type="button" class="btn btn-default" data-dismiss="modal" ng-disabled="apiMiddleware.apiHandler.inprocess">{{"cancel" | translate}}</button>\r\n                  <button type="submit" class="btn btn-primary" ng-disabled="apiMiddleware.apiHandler.inprocess">{{\'compress\' | translate}}</button>\r\n              </div>\r\n            </div>\r\n        </form>\r\n    </div>\r\n  </div>\r\n</div>\r\n\r\n<div class="modal animated fadeIn" id="extract" ng-init="singleSelection().emptyName()">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n        <form ng-submit="extract()">\r\n            <div class="modal-header">\r\n              <button type="button" class="close" data-dismiss="modal">\r\n                  <span aria-hidden="true">&times;</span>\r\n                  <span class="sr-only">{{"close" | translate}}</span>\r\n              </button>\r\n              <h4 class="modal-title">{{\'extract_item\' | translate}}</h4>\r\n            </div>\r\n            <div class="modal-body">\r\n              <div ng-show="apiMiddleware.apiHandler.asyncSuccess">\r\n                  <div class="label label-success error-msg">{{\'extraction_started\' | translate}}</div>\r\n              </div>\r\n              <div ng-hide="apiMiddleware.apiHandler.asyncSuccess">\r\n                  <label class="radio">{{\'enter_folder_name_for_extraction\' | translate}} <b>{{singleSelection().model.name}}</b></label>\r\n                  <input class="form-control" ng-model="singleSelection().tempModel.name" autofocus="autofocus">\r\n              </div>\r\n              <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n            </div>\r\n            <div class="modal-footer">\r\n              <div ng-show="apiMiddleware.apiHandler.asyncSuccess">\r\n                  <button type="button" class="btn btn-default" data-dismiss="modal" ng-disabled="apiMiddleware.apiHandler.inprocess">{{"close" | translate}}</button>\r\n              </div>\r\n              <div ng-hide="apiMiddleware.apiHandler.asyncSuccess">\r\n                  <button type="button" class="btn btn-default" data-dismiss="modal" ng-disabled="apiMiddleware.apiHandler.inprocess">{{"cancel" | translate}}</button>\r\n                  <button type="submit" class="btn btn-primary" ng-disabled="apiMiddleware.apiHandler.inprocess">{{\'extract\' | translate}}</button>\r\n              </div>\r\n            </div>\r\n        </form>\r\n    </div>\r\n  </div>\r\n</div>\r\n\r\n<div class="modal animated fadeIn" id="edit" ng-class="{\'modal-fullscreen\': fullscreen}">\r\n  <div class="modal-dialog modal-lg">\r\n    <div class="modal-content">\r\n        <form ng-submit="edit()">\r\n            <div class="modal-header">\r\n              <button type="button" class="close" data-dismiss="modal">\r\n                  <span aria-hidden="true">&times;</span>\r\n                  <span class="sr-only">{{"close" | translate}}</span>\r\n              </button>\r\n              <button type="button" class="close fullscreen" ng-click="fullscreen=!fullscreen">\r\n                  <i class="glyphicon glyphicon-fullscreen"></i>\r\n                  <span class="sr-only">{{\'toggle_fullscreen\' | translate}}</span>\r\n              </button>\r\n              <h4 class="modal-title">{{\'edit_file\' | translate}}</h4>\r\n            </div>\r\n            <div class="modal-body">\r\n                <label class="radio bold">{{ singleSelection().model.fullPath() }}</label>\r\n                <span class="label label-warning" ng-show="apiMiddleware.apiHandler.inprocess">{{\'loading\' | translate}} ...</span>\r\n                <textarea class="form-control code" ng-model="singleSelection().tempModel.content" ng-show="!apiMiddleware.apiHandler.inprocess" autofocus="autofocus"></textarea>\r\n                <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n            </div>\r\n            <div class="modal-footer">\r\n              <button type="button" class="btn btn-default" data-dismiss="modal" ng-disabled="apiMiddleware.apiHandler.inprocess">{{\'close\' | translate}}</button>\r\n              <button type="submit" class="btn btn-primary" ng-show="config.allowedActions.edit" ng-disabled="apiMiddleware.apiHandler.inprocess">保存</button>\r\n            </div>\r\n        </form>\r\n    </div>\r\n  </div>\r\n</div>\r\n\r\n<div class="modal animated fadeIn" id="newfolder">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n        <form ng-submit="createFolder()">\r\n            <div class="modal-header">\r\n              <button type="button" class="close" data-dismiss="modal">\r\n                  <span aria-hidden="true">&times;</span>\r\n                  <span class="sr-only">{{"close" | translate}}</span>\r\n              </button>\r\n              <h4 class="modal-title">{{\'new_folder\' | translate}}</h4>\r\n            </div>\r\n            <div class="modal-body">\r\n              <label class="radio">{{\'folder_name\' | translate}}</label>\r\n              <input class="form-control" ng-model="singleSelection().tempModel.name" autofocus="autofocus">\r\n              <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n            </div>\r\n            <div class="modal-footer">\r\n              <button type="button" class="btn btn-default" data-dismiss="modal" ng-disabled="apiMiddleware.apiHandler.inprocess">{{"cancel" | translate}}</button>\r\n              <button type="submit" class="btn btn-primary" ng-disabled="apiMiddleware.apiHandler.inprocess">{{\'create\' | translate}}</button>\r\n            </div>\r\n        </form>\r\n    </div>\r\n  </div>\r\n</div>\r\n\r\n<div class="modal animated fadeIn" id="uploadfile">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n        <form>\r\n            <div class="modal-header">\r\n              <button type="button" class="close" data-dismiss="modal">\r\n                  <span aria-hidden="true">&times;</span>\r\n                  <span class="sr-only">{{"close" | translate}}</span>\r\n              </button>\r\n              <h4 class="modal-title">{{"upload_files" | translate}}</h4>\r\n            </div>\r\n            <div class="modal-body">\r\n              <label class="radio">\r\n                {{"files_will_uploaded_to" | translate}} \r\n                <b>/{{fileNavigator.currentPath.join(\'/\')}}</b>\r\n              </label>\r\n              <button class="btn btn-default btn-block" ngf-select="$parent.addForUpload($files)" ngf-multiple="true">\r\n                {{"select_files" | translate}}\r\n              </button>\r\n              \r\n              <div class="upload-list">\r\n                <ul class="list-group">\r\n                  <li class="list-group-item" ng-repeat="(index, uploadFile) in $parent.uploadFileList">\r\n                    <button class="btn btn-sm btn-danger pull-right" ng-click="$parent.removeFromUpload(index)">\r\n                        &times;\r\n                    </button>\r\n                    <h5 class="list-group-item-heading">{{uploadFile.name}}</h5>\r\n                    <p class="list-group-item-text">{{uploadFile.size | humanReadableFileSize}}</p>\r\n                  </li>\r\n                </ul>\r\n                <div ng-show="apiMiddleware.apiHandler.inprocess">\r\n                  <em>{{"uploading" | translate}}... {{apiMiddleware.apiHandler.progress}}%</em>\r\n                  <div class="progress mb0">\r\n                    <div class="progress-bar active" role="progressbar" aria-valuenow="{{apiMiddleware.apiHandler.progress}}" aria-valuemin="0" aria-valuemax="100" style="width: {{apiMiddleware.apiHandler.progress}}%"></div>\r\n                  </div>\r\n                </div>\r\n              </div>\r\n              <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n            </div>\r\n            <div class="modal-footer">\r\n              <div>\r\n                  <button type="button" class="btn btn-default" data-dismiss="modal">{{"cancel" | translate}}</button>\r\n                  <button type="submit" class="btn btn-primary" ng-disabled="!$parent.uploadFileList.length || apiMiddleware.apiHandler.inprocess" ng-click="uploadFiles()">{{\'upload\' | translate}}</button>\r\n              </div>\r\n            </div>\r\n        </form>\r\n    </div>\r\n  </div>\r\n</div>\r\n\r\n<div class="modal animated fadeIn" id="changepermissions">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n        <form ng-submit="changePermissions()">\r\n            <div class="modal-header">\r\n              <button type="button" class="close" data-dismiss="modal">\r\n                  <span aria-hidden="true">&times;</span>\r\n                  <span class="sr-only">{{"close" | translate}}</span>\r\n              </button>\r\n              <h4 class="modal-title">{{\'change_permissions\' | translate}}</h4>\r\n            </div>\r\n            <div class="modal-body">\r\n              <table class="table mb0">\r\n                  <thead>\r\n                      <tr>\r\n                          <th>{{\'permissions\' | translate}}</th>\r\n                          <th class="col-xs-1 text-center">{{\'read\' | translate}}</th>\r\n                          <th class="col-xs-1 text-center">{{\'write\' | translate}}</th>\r\n                          <th class="col-xs-1 text-center">{{\'exec\' | translate}}</th>\r\n                      </tr>\r\n                  </thead>\r\n                  <tbody>\r\n                      <tr ng-repeat="(permTypeKey, permTypeValue) in temp.tempModel.perms">\r\n                          <td>{{permTypeKey | translate}}</td>\r\n                          <td ng-repeat="(permKey, permValue) in permTypeValue" class="col-xs-1 text-center" ng-click="main()">\r\n                              <label class="col-xs-12">\r\n                                <input type="checkbox" ng-model="temp.tempModel.perms[permTypeKey][permKey]">\r\n                              </label>\r\n                          </td>\r\n                      </tr>\r\n                </tbody>\r\n              </table>\r\n              <div class="checkbox" ng-show="config.enablePermissionsRecursive && selectionHas(\'dir\')">\r\n                <label>\r\n                  <input type="checkbox" ng-model="temp.tempModel.recursive"> {{\'recursive\' | translate}}\r\n                </label>\r\n              </div>\r\n              <div class="clearfix mt10">\r\n                  <span class="label label-primary pull-left" ng-hide="temp.multiple">\r\n                    {{\'original\' | translate}}: \r\n                    {{temp.model.perms.toCode(selectionHas(\'dir\') ? \'d\':\'-\')}} \r\n                    ({{temp.model.perms.toOctal()}})\r\n                  </span>\r\n                  <span class="label label-primary pull-right">\r\n                    {{\'changes\' | translate}}: \r\n                    {{temp.tempModel.perms.toCode(selectionHas(\'dir\') ? \'d\':\'-\')}} \r\n                    ({{temp.tempModel.perms.toOctal()}})\r\n                  </span>\r\n              </div>\r\n              <div ng-include data-src="\'error-bar\'" class="clearfix"></div>\r\n            </div>\r\n            <div class="modal-footer">\r\n              <button type="button" class="btn btn-default" data-dismiss="modal">{{"cancel" | translate}}</button>\r\n              <button type="submit" class="btn btn-primary" ng-disabled="">{{\'change\' | translate}}</button>\r\n            </div>\r\n        </form>\r\n    </div>\r\n  </div>\r\n</div>\r\n\r\n<div class="modal animated fadeIn" id="selector" ng-controller="ModalFileManagerCtrl">\r\n  <div class="modal-dialog">\r\n    <div class="modal-content">\r\n      <div class="modal-header">\r\n        <button type="button" class="close" data-dismiss="modal">\r\n            <span aria-hidden="true">&times;</span>\r\n            <span class="sr-only">{{"close" | translate}}</span>\r\n        </button>\r\n        <h4 class="modal-title">{{"select_destination_folder" | translate}}</h4>\r\n      </div>\r\n      <div class="modal-body">\r\n        <div>\r\n            <div ng-include="config.tplPath + \'/current-folder-breadcrumb.html\'"></div>\r\n            <div ng-include="config.tplPath + \'/main-table-modal.html\'"></div>\r\n            <hr />\r\n            <button class="btn btn-sm btn-default" ng-click="selectCurrent()">\r\n                <i class="glyphicon"></i> {{"select_this" | translate}}\r\n            </button>\r\n        </div>\r\n      </div>\r\n      <div class="modal-footer">\r\n        <button type="button" class="btn btn-default" data-dismiss="modal" ng-disabled="apiMiddleware.apiHandler.inprocess">{{"close" | translate}}</button>\r\n      </div>\r\n    </div>\r\n  </div>\r\n</div>\r\n\r\n<script type="text/ng-template" id="path-selector">\r\n  <div class="panel panel-primary mt10 mb0">\r\n    <div class="panel-body">\r\n        <div class="detail-sources">\r\n          <div class="like-code mr5"><b>{{"selection" | translate}}:</b>\r\n            <span ng-include="\'selected-files-msg\'"></span>\r\n          </div>\r\n        </div>\r\n        <div class="detail-sources">\r\n          <div class="like-code mr5">\r\n            <b>{{"destination" | translate}}:</b> {{ getSelectedPath() }}\r\n          </div>\r\n          <a href="" class="label label-primary" ng-click="openNavigator(fileNavigator.currentPath)">\r\n            {{\'change\' | translate}}\r\n          </a>\r\n        </div>\r\n    </div>\r\n  </div>\r\n</script>\r\n\r\n<script type="text/ng-template" id="error-bar">\r\n  <div class="label label-danger error-msg pull-left animated fadeIn" ng-show="apiMiddleware.apiHandler.error">\r\n    <i class="glyphicon glyphicon-remove-circle"></i>\r\n    <span>{{apiMiddleware.apiHandler.error}}</span>\r\n  </div>\r\n</script>\r\n\r\n<script type="text/ng-template" id="selected-files-msg">\r\n  <span ng-show="temps.length == 1">\r\n    {{singleSelection().model.name}}\r\n  </span>\r\n  <span ng-show="temps.length > 1">\r\n    {{\'these_elements\' | translate:totalSelecteds()}}\r\n    <a href="" class="label label-primary" ng-click="showDetails = !showDetails">\r\n      {{showDetails ? \'-\' : \'+\'}} {{\'details\' | translate}}\r\n    </a>\r\n  </span>\r\n  <div ng-show="temps.length > 1 &amp;&amp; showDetails">\r\n    <ul class="selected-file-details">\r\n      <li ng-repeat="tempItem in temps">\r\n        <b>{{tempItem.model.name}}</b>\r\n      </li>\r\n    </ul>\r\n  </div>\r\n</script>\r\n'),
            e.put("src/templates/navbar.html", '<nav class="navbar navbar-inverse lefts">\r\n    <div class="container-fluid">\r\n        <div class="row">\r\n            <div class="col-sm-9 col-md-9 hidden-xs">\r\n                <div ng-show="!config.breadcrumb">\r\n                    <a class="navbar-brand hidden-xs ng-binding" href="">angular-{{"filemanager" | translate}}</a>\r\n                </div>\r\n                <div ng-include="config.tplPath + \'/current-folder-breadcrumb.html\'" ng-show="config.breadcrumb">\r\n                </div>\r\n            </div>\r\n            <div class="col-sm-3 col-md-3">\r\n     <div class="navbar-collapse">\r\n                    <div class="navbar-form navbar-right text-right">\r\n                        <div class="pull-left visible-xs" ng-if="fileNavigator.currentPath.length">\r\n  <button class="btn btn-primary btn-flat" ng-click="fileNavigator.upDir()">\r\n                                <i class="glyphicon glyphicon-chevron-left"></i>\r\n                            </button>\r\n                            {{fileNavigator.getCurrentFolderName() | strLimit : 12}}\r\n</div><div class="btn-group">\r\n                            <button class="btn btn-flat btn-sm dropdown-toggle" type="button" id="dropDownMenuSearch" data-toggle="dropdown" aria-expanded="true">\r\n                                <i class="glyphicon glyphicon-search mr2"></i>\r\n                            </button>\r\n                            <div class="dropdown-menu animated fast fadeIn pull-right" role="menu" aria-labelledby="dropDownMenuLang">\r\n                                <input type="text" class="form-control indent" ng-show="config.searchForm" placeholder="{{\'search\' | translate}}..." ng-model="$parent.query">\r\n                            </div>\r\n                        </div>\r\n\r\n                        <button class="btn btn-flat btn-sm" ng-click="$parent.setTemplate(\'main-icons.html\')" ng-show="$parent.viewTemplate !==\'main-icons.html\'" title="{{\'icons\' | translate}}">\r\n                            <i class="glyphicon glyphicon-th-large"></i>\r\n                        </button>\r\n\r\n                        <button class="btn btn-flat btn-sm" ng-click="$parent.setTemplate(\'main-table.html\')" ng-show="$parent.viewTemplate !==\'main-table.html\'" title="{{\'list\' | translate}}">\r\n                            <i class="glyphicon glyphicon-th-list"></i>\r\n                        </button>\r\n\r\n                       \r\n\r\n             </div></div></div> </div>\r\n</nav>\r\n'),
            e.put("src/templates/sidebar.html", '<ul class="nav nav-sidebar file-tree-root">\r\n    <li ng-repeat="item in fileNavigator.history" ng-include="\'folder-branch-item\'" ng-class="{\'active\': item.name == fileNavigator.currentPath.join(\'/\')}"></li>\r\n</ul>\r\n\r\n<script type="text/ng-template" id="folder-branch-item">\r\n    <a href="" ng-click="fileNavigator.folderClick(item.item)" class="animated fast fadeInDown">\r\n\r\n        <span class="point">\r\n            <i class="glyphicon glyphicon-chevron-down" ng-show="isInThisPath(item.name)"></i>\r\n            <i class="glyphicon glyphicon-chevron-right" ng-show="!isInThisPath(item.name)"></i>\r\n        </span>\r\n\r\n        <i class="glyphicon glyphicon-folder-open mr2" ng-show="isInThisPath(item.name)"></i>\r\n        <i class="glyphicon glyphicon-folder-close mr2" ng-show="!isInThisPath(item.name)"></i>\r\n        {{ (item.name.split(\'/\').pop() || fileNavigator.getBasePath().join(\'/\') || \'/\') | strLimit : 30 }}\r\n    </a>\r\n    <ul class="nav nav-sidebar">\r\n        <li ng-repeat="item in item.nodes" ng-include="\'folder-branch-item\'" ng-class="{\'active\': item.name == fileNavigator.currentPath.join(\'/\')}"></li>\r\n    </ul>\r\n<\/script>'),
            e.put("src/templates/spinner.html", '<div class="spinner-wrapper col-xs-12">\r\n    <svg class="spinner-container" style="width:65px;height:65px" viewBox="0 0 44 44">\r\n        <circle class="path" cx="22" cy="22" r="20" fill="none" stroke-width="4"></circle>\r\n    </svg>\r\n</div>')
    }
]);