// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package webdav provides a WebDAV server implementation.
package webdav // import "golang.org/x/net/webdav"

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	model "github.com/cloudreve/Cloudreve/v3/models"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/driver/local"
	"github.com/cloudreve/Cloudreve/v3/pkg/filesystem/fsctx"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
)

type Handler struct {
	// Prefix is the URL path prefix to strip from WebDAV resource paths.
	Prefix string
	// LockSystem is the lock management system.
	LockSystem map[uint]LockSystem
	// Logger is an optional error logger. If non-nil, it will be called
	// for all HTTP requests.
	Logger func(*http.Request, error)
}

func (h *Handler) stripPrefix(p string, uid uint) (string, int, error) {
	if h.Prefix == "" {
		return p, http.StatusOK, nil
	}
	prefix := h.Prefix
	if r := strings.TrimPrefix(p, prefix); len(r) < len(p) {
		if len(r) == 0 {
			r = "/"
		}
		return util.RemoveSlash(r), http.StatusOK, nil
	}
	return p, http.StatusNotFound, errPrefixMismatch
}

// isPathExist 路径是否存在
func isPathExist(ctx context.Context, fs *filesystem.FileSystem, path string) (bool, FileInfo) {
	// 尝试目录
	if ok, folder := fs.IsPathExist(path); ok {
		return ok, folder
	}
	if ok, file := fs.IsFileExist(path); ok {
		return ok, file
	}
	return false, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request, fs *filesystem.FileSystem) {
	status, err := http.StatusBadRequest, errUnsupportedMethod
	if h.LockSystem == nil {
		status, err = http.StatusInternalServerError, errNoLockSystem
	} else {
		// 检查并新建LockSystem
		if _, ok := h.LockSystem[fs.User.ID]; !ok {
			h.LockSystem[fs.User.ID] = NewMemLS()
		}
		switch r.Method {
		case "OPTIONS":
			status, err = h.handleOptions(w, r, fs)
		case "GET", "HEAD", "POST":
			status, err = h.handleGetHeadPost(w, r, fs)
		case "DELETE":
			status, err = h.handleDelete(w, r, fs)
		case "PUT":
			status, err = h.handlePut(w, r, fs)
		case "MKCOL":
			status, err = h.handleMkcol(w, r, fs)
		case "COPY", "MOVE":
			status, err = h.handleCopyMove(w, r, fs)
		case "LOCK":
			status, err = h.handleLock(w, r, fs)
		case "UNLOCK":
			status, err = h.handleUnlock(w, r, fs)
		case "PROPFIND":
			status, err = h.handlePropfind(w, r, fs)
		case "PROPPATCH":
			status, err = h.handleProppatch(w, r, fs)
		}
	}

	if status != 0 {
		w.WriteHeader(status)
		if status != http.StatusNoContent {
			w.Write([]byte(StatusText(status)))
		}
	}
	if h.Logger != nil {
		h.Logger(r, err)
	}
}

// OK
func (h *Handler) lock(now time.Time, root string, fs *filesystem.FileSystem) (token string, status int, err error) {
	token, err = h.LockSystem[fs.User.ID].Create(now, LockDetails{
		Root:      root,
		Duration:  infiniteTimeout,
		ZeroDepth: true,
	})
	if err != nil {
		if err == ErrLocked {
			return "", StatusLocked, err
		}
		return "", http.StatusInternalServerError, err
	}
	return token, 0, nil
}

// ok
func (h *Handler) confirmLocks(r *http.Request, src, dst string, fs *filesystem.FileSystem) (release func(), status int, err error) {
	hdr := r.Header.Get("If")
	if hdr == "" {
		// An empty If header means that the client hasn't previously created locks.
		// Even if this client doesn't care about locks, we still need to check that
		// the resources aren't locked by another client, so we create temporary
		// locks that would conflict with another client's locks. These temporary
		// locks are unlocked at the end of the HTTP request.
		now, srcToken, dstToken := time.Now(), "", ""
		if src != "" {
			srcToken, status, err = h.lock(now, src, fs)
			if err != nil {
				return nil, status, err
			}
		}
		if dst != "" {
			dstToken, status, err = h.lock(now, dst, fs)
			if err != nil {
				if srcToken != "" {
					h.LockSystem[fs.User.ID].Unlock(now, srcToken)
				}
				return nil, status, err
			}
		}

		return func() {
			if dstToken != "" {
				h.LockSystem[fs.User.ID].Unlock(now, dstToken)
			}
			if srcToken != "" {
				h.LockSystem[fs.User.ID].Unlock(now, srcToken)
			}
		}, 0, nil
	}

	ih, ok := parseIfHeader(hdr)
	if !ok {
		return nil, http.StatusBadRequest, errInvalidIfHeader
	}
	// ih is a disjunction (OR) of ifLists, so any ifList will do.
	for _, l := range ih.lists {
		lsrc := l.resourceTag
		if lsrc == "" {
			lsrc = src
		} else {
			u, err := url.Parse(lsrc)
			if err != nil {
				continue
			}
			//if u.Host != r.Host {
			//	continue
			//}
			lsrc, status, err = h.stripPrefix(u.Path, fs.User.ID)
			if err != nil {
				return nil, status, err
			}
		}
		release, err = h.LockSystem[fs.User.ID].Confirm(
			time.Now(),
			lsrc,
			dst,
			l.conditions...,
		)
		if err == ErrConfirmationFailed {
			continue
		}
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		return release, 0, nil
	}
	// Section 10.4.1 says that "If this header is evaluated and all state lists
	// fail, then the request must fail with a 412 (Precondition Failed) status."
	// We follow the spec even though the cond_put_corrupt_token test case from
	// the litmus test warns on seeing a 412 instead of a 423 (Locked).
	return nil, http.StatusPreconditionFailed, ErrLocked
}

//OK
func (h *Handler) handleOptions(w http.ResponseWriter, r *http.Request, fs *filesystem.FileSystem) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path, fs.User.ID)
	if err != nil {
		return status, err
	}
	ctx := r.Context()
	allow := "OPTIONS, LOCK, PUT, MKCOL"
	if exist, fi := isPathExist(ctx, fs, reqPath); exist {
		if fi.IsDir() {
			allow = "OPTIONS, LOCK, DELETE, PROPPATCH, COPY, MOVE, UNLOCK, PROPFIND"
		} else {
			allow = "OPTIONS, LOCK, GET, HEAD, POST, DELETE, PROPPATCH, COPY, MOVE, UNLOCK, PROPFIND, PUT"
		}
	}
	w.Header().Set("Allow", allow)
	// http://www.webdav.org/specs/rfc4918.html#dav.compliance.classes
	w.Header().Set("DAV", "1, 2")
	// http://msdn.microsoft.com/en-au/library/cc250217.aspx
	w.Header().Set("MS-Author-Via", "DAV")
	return 0, nil
}

// OK
func (h *Handler) handleGetHeadPost(w http.ResponseWriter, r *http.Request, fs *filesystem.FileSystem) (status int, err error) {
	defer fs.Recycle()

	reqPath, status, err := h.stripPrefix(r.URL.Path, fs.User.ID)
	if err != nil {
		return status, err
	}

	ctx := r.Context()

	exist, file := fs.IsFileExist(reqPath)
	if !exist {
		return http.StatusNotFound, nil
	}
	fs.SetTargetFile(&[]model.File{*file})

	rs, err := fs.Preview(ctx, 0, false)
	if err != nil {
		if err == filesystem.ErrObjectNotExist {
			return http.StatusNotFound, err
		}
		return http.StatusInternalServerError, err
	}

	etag, err := findETag(ctx, fs, h.LockSystem[fs.User.ID], reqPath, &fs.FileTarget[0])
	if err != nil {
		return http.StatusInternalServerError, err
	}
	w.Header().Set("ETag", etag)

	if !rs.Redirect {
		defer rs.Content.Close()
		// 获取文件内容
		http.ServeContent(w, r, reqPath, fs.FileTarget[0].UpdatedAt, rs.Content)
		return 0, nil
	}

	http.Redirect(w, r, rs.URL, 301)

	return 0, nil
}

// OK
func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request, fs *filesystem.FileSystem) (status int, err error) {
	defer fs.Recycle()

	reqPath, status, err := h.stripPrefix(r.URL.Path, fs.User.ID)
	if err != nil {
		return status, err
	}
	release, status, err := h.confirmLocks(r, reqPath, "", fs)
	if err != nil {
		return status, err
	}
	defer release()

	ctx := r.Context()

	// 尝试作为文件删除
	if ok, file := fs.IsFileExist(reqPath); ok {
		if err := fs.Delete(ctx, []uint{}, []uint{file.ID}, false); err != nil {
			return http.StatusMethodNotAllowed, err
		}
		return http.StatusNoContent, nil
	}

	// 尝试作为目录删除
	if ok, folder := fs.IsPathExist(reqPath); ok {
		if err := fs.Delete(ctx, []uint{folder.ID}, []uint{}, false); err != nil {
			return http.StatusMethodNotAllowed, err
		}
		return http.StatusNoContent, nil
	}

	return http.StatusNotFound, nil
}

// OK
func (h *Handler) handlePut(w http.ResponseWriter, r *http.Request, fs *filesystem.FileSystem) (status int, err error) {
	reqPath, status, err := h.stripPrefix(r.URL.Path, fs.User.ID)
	if err != nil {
		return status, err
	}
	release, status, err := h.confirmLocks(r, reqPath, "", fs)
	if err != nil {
		return status, err
	}
	defer release()
	// TODO(rost): Support the If-Match, If-None-Match headers? See bradfitz'
	// comments in http.checkEtag.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = context.WithValue(ctx, fsctx.HTTPCtx, r.Context())
	ctx = context.WithValue(ctx, fsctx.CancelFuncCtx, cancel)
	ctx = context.WithValue(ctx, fsctx.ValidateCapacityOnceCtx, &sync.Once{})

	fileSize, err := strconv.ParseUint(r.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return http.StatusMethodNotAllowed, err
	}
	fileName := path.Base(reqPath)
	filePath := path.Dir(reqPath)
	fileData := local.FileStream{
		MIMEType:    r.Header.Get("Content-Type"),
		File:        r.Body,
		Size:        fileSize,
		Name:        fileName,
		VirtualPath: filePath,
	}

	// 判断文件是否已存在
	exist, originFile := fs.IsFileExist(reqPath)
	if exist {
		// 已存在，为更新操作

		// 检查此文件是否有软链接
		fileList, err := model.RemoveFilesWithSoftLinks([]model.File{*originFile})
		if err == nil && len(fileList) == 0 {
			// 如果包含软连接，应重新生成新文件副本，并更新source_name
			originFile.SourceName = fs.GenerateSavePath(ctx, fileData)
			fs.Use("AfterUpload", filesystem.HookUpdateSourceName)
			fs.Use("AfterUploadCanceled", filesystem.HookUpdateSourceName)
			fs.Use("AfterValidateFailed", filesystem.HookUpdateSourceName)
		}

		fs.Use("BeforeUpload", filesystem.HookResetPolicy)
		fs.Use("BeforeUpload", filesystem.HookValidateFile)
		fs.Use("BeforeUpload", filesystem.HookChangeCapacity)
		fs.Use("AfterUploadCanceled", filesystem.HookCleanFileContent)
		fs.Use("AfterUploadCanceled", filesystem.HookClearFileSize)
		fs.Use("AfterUploadCanceled", filesystem.HookGiveBackCapacity)
		fs.Use("AfterUploadCanceled", filesystem.HookCancelContext)
		fs.Use("AfterUpload", filesystem.GenericAfterUpdate)
		fs.Use("AfterValidateFailed", filesystem.HookCleanFileContent)
		fs.Use("AfterValidateFailed", filesystem.HookClearFileSize)
		fs.Use("AfterValidateFailed", filesystem.HookGiveBackCapacity)
		ctx = context.WithValue(ctx, fsctx.FileModelCtx, *originFile)
	} else {
		// 给文件系统分配钩子
		fs.Use("BeforeUpload", filesystem.HookValidateFile)
		fs.Use("BeforeUpload", filesystem.HookValidateCapacity)
		fs.Use("AfterUploadCanceled", filesystem.HookDeleteTempFile)
		fs.Use("AfterUploadCanceled", filesystem.HookGiveBackCapacity)
		fs.Use("AfterUploadCanceled", filesystem.HookCancelContext)
		fs.Use("AfterUpload", filesystem.GenericAfterUpload)
		fs.Use("AfterValidateFailed", filesystem.HookDeleteTempFile)
		fs.Use("AfterValidateFailed", filesystem.HookGiveBackCapacity)
		fs.Use("AfterUploadFailed", filesystem.HookGiveBackCapacity)
	}

	// 执行上传
	err = fs.Upload(ctx, fileData)
	if err != nil {
		return http.StatusMethodNotAllowed, err
	}

	etag, err := findETag(ctx, fs, h.LockSystem[fs.User.ID], reqPath, &fs.FileTarget[0])
	if err != nil {
		return http.StatusInternalServerError, err
	}
	w.Header().Set("ETag", etag)
	return http.StatusCreated, nil
}

// OK
func (h *Handler) handleMkcol(w http.ResponseWriter, r *http.Request, fs *filesystem.FileSystem) (status int, err error) {
	defer fs.Recycle()

	reqPath, status, err := h.stripPrefix(r.URL.Path, fs.User.ID)
	if err != nil {
		return status, err
	}
	release, status, err := h.confirmLocks(r, reqPath, "", fs)
	if err != nil {
		return status, err
	}
	defer release()

	ctx := r.Context()

	if r.ContentLength > 0 {
		return http.StatusUnsupportedMediaType, nil
	}
	if strings.Contains(r.UserAgent(), "rclone") {
		if _, ok := ctx.Value(fsctx.IgnoreConflictCtx).(bool); !ok {
			ctx = context.WithValue(ctx, fsctx.IgnoreConflictCtx, true)
		}
	}
	if _, err := fs.CreateDirectory(ctx, reqPath); err != nil {
		return http.StatusConflict, err
	}
	return http.StatusCreated, nil
}

// OK
func (h *Handler) handleCopyMove(w http.ResponseWriter, r *http.Request, fs *filesystem.FileSystem) (status int, err error) {
	defer fs.Recycle()

	hdr := r.Header.Get("Destination")
	if hdr == "" {
		return http.StatusBadRequest, errInvalidDestination
	}
	u, err := url.Parse(hdr)
	if err != nil {
		return http.StatusBadRequest, errInvalidDestination
	}
	//if u.Host != "" && u.Host != r.Host {
	//	return http.StatusBadGateway, errInvalidDestination
	//}

	src, status, err := h.stripPrefix(r.URL.Path, fs.User.ID)
	if err != nil {
		return status, err
	}

	dst, status, err := h.stripPrefix(u.Path, fs.User.ID)
	if err != nil {
		return status, err
	}

	if dst == "" {
		return http.StatusBadGateway, errInvalidDestination
	}
	if dst == src {
		return http.StatusForbidden, errDestinationEqualsSource
	}

	ctx := r.Context()

	isExist, target := isPathExist(ctx, fs, src)

	if !isExist {
		return http.StatusNotFound, nil
	}

	if r.Method == "COPY" {
		// Section 7.5.1 says that a COPY only needs to lock the destination,
		// not both destination and source. Strictly speaking, this is racy,
		// even though a COPY doesn't modify the source, if a concurrent
		// operation modifies the source. However, the litmus test explicitly
		// checks that COPYing a locked-by-another source is OK.
		release, status, err := h.confirmLocks(r, "", dst, fs)
		if err != nil {
			return status, err
		}
		defer release()

		// Section 9.8.3 says that "The COPY method on a collection without a Depth
		// header must act as if a Depth header with value "infinity" was included".
		depth := infiniteDepth
		if hdr := r.Header.Get("Depth"); hdr != "" {
			depth = parseDepth(hdr)
			if depth != 0 && depth != infiniteDepth {
				// Section 9.8.3 says that "A client may submit a Depth header on a
				// COPY on a collection with a value of "0" or "infinity"."
				return http.StatusBadRequest, errInvalidDepth
			}
		}
		return copyFiles(ctx, fs, target, dst, r.Header.Get("Overwrite") != "F", depth, 0)
	}

	// windows下，某些情况下（网盘根目录下）Office保存文件时附带的锁token只包含源文件，
	// 此处暂时去除了对dst锁的检查
	release, status, err := h.confirmLocks(r, src, "", fs)
	if err != nil {
		return status, err
	}
	defer release()

	// Section 9.9.2 says that "The MOVE method on a collection must act as if
	// a "Depth: infinity" header was used on it. A client must not submit a
	// Depth header on a MOVE on a collection with any value but "infinity"."
	if hdr := r.Header.Get("Depth"); hdr != "" {
		if parseDepth(hdr) != infiniteDepth {
			return http.StatusBadRequest, errInvalidDepth
		}
	}
	return moveFiles(ctx, fs, target, dst, r.Header.Get("Overwrite") == "T")
}

// OK
func (h *Handler) handleLock(w http.ResponseWriter, r *http.Request, fs *filesystem.FileSystem) (retStatus int, retErr error) {
	defer fs.Recycle()

	duration, err := parseTimeout(r.Header.Get("Timeout"))
	if err != nil {
		return http.StatusBadRequest, err
	}
	li, status, err := readLockInfo(r.Body)
	if err != nil {
		return status, err
	}

	//ctx := r.Context()
	token, ld, now, created := "", LockDetails{}, time.Now(), false
	if li == (lockInfo{}) {
		// An empty lockInfo means to refresh the lock.
		ih, ok := parseIfHeader(r.Header.Get("If"))
		if !ok {
			return http.StatusBadRequest, errInvalidIfHeader
		}
		if len(ih.lists) == 1 && len(ih.lists[0].conditions) == 1 {
			token = ih.lists[0].conditions[0].Token
		}
		if token == "" {
			return http.StatusBadRequest, errInvalidLockToken
		}
		ld, err = h.LockSystem[fs.User.ID].Refresh(now, token, duration)
		if err != nil {
			if err == ErrNoSuchLock {
				return http.StatusPreconditionFailed, err
			}
			return http.StatusInternalServerError, err
		}

	} else {
		// Section 9.10.3 says that "If no Depth header is submitted on a LOCK request,
		// then the request MUST act as if a "Depth:infinity" had been submitted."
		depth := infiniteDepth
		if hdr := r.Header.Get("Depth"); hdr != "" {
			depth = parseDepth(hdr)
			if depth != 0 && depth != infiniteDepth {
				// Section 9.10.3 says that "Values other than 0 or infinity must not be
				// used with the Depth header on a LOCK method".
				return http.StatusBadRequest, errInvalidDepth
			}
		}
		reqPath, status, err := h.stripPrefix(r.URL.Path, fs.User.ID)
		if err != nil {
			return status, err
		}
		ld = LockDetails{
			Root:      reqPath,
			Duration:  duration,
			OwnerXML:  li.Owner.InnerXML,
			ZeroDepth: depth == 0,
		}
		token, err = h.LockSystem[fs.User.ID].Create(now, ld)
		if err != nil {
			if err == ErrLocked {
				return StatusLocked, err
			}
			return http.StatusInternalServerError, err
		}
		defer func() {
			if retErr != nil {
				h.LockSystem[fs.User.ID].Unlock(now, token)
			}
		}()

		// Create the resource if it didn't previously exist.
		//if _, err := h.FileSystem.Stat(ctx, reqPath); err != nil {
		//	f, err := h.FileSystem.OpenFile(ctx, reqPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		//	if err != nil {
		//		// TODO: detect missing intermediate dirs and return http.StatusConflict?
		//		return http.StatusInternalServerError, err
		//	}
		//	f.Close()
		//	created = true
		//}

		// http://www.webdav.org/specs/rfc4918.html#HEADER_Lock-Token says that the
		// Lock-Token value is a Coded-URL. We add angle brackets.
		w.Header().Set("Lock-Token", "<"+token+">")
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	if created {
		// This is "w.WriteHeader(http.StatusCreated)" and not "return
		// http.StatusCreated, nil" because we write our own (XML) response to w
		// and Handler.ServeHTTP would otherwise write "Created".
		w.WriteHeader(http.StatusCreated)
	}
	writeLockInfo(w, token, ld)
	return 0, nil
}

// OK
func (h *Handler) handleUnlock(w http.ResponseWriter, r *http.Request, fs *filesystem.FileSystem) (status int, err error) {
	defer fs.Recycle()

	// http://www.webdav.org/specs/rfc4918.html#HEADER_Lock-Token says that the
	// Lock-Token value is a Coded-URL. We strip its angle brackets.
	t := r.Header.Get("Lock-Token")
	if len(t) < 2 || t[0] != '<' || t[len(t)-1] != '>' {
		return http.StatusBadRequest, errInvalidLockToken
	}
	t = t[1 : len(t)-1]

	switch err = h.LockSystem[fs.User.ID].Unlock(time.Now(), t); err {
	case nil:
		return http.StatusNoContent, err
	case ErrForbidden:
		return http.StatusForbidden, err
	case ErrLocked:
		return StatusLocked, err
	case ErrNoSuchLock:
		return http.StatusConflict, err
	default:
		return http.StatusInternalServerError, err
	}
}

// OK
func (h *Handler) handlePropfind(w http.ResponseWriter, r *http.Request, fs *filesystem.FileSystem) (status int, err error) {
	defer fs.Recycle()

	reqPath, status, err := h.stripPrefix(r.URL.Path, fs.User.ID)
	if err != nil {
		return status, err
	}
	ctx := r.Context()
	ok, fi := isPathExist(ctx, fs, reqPath)
	if !ok {
		return http.StatusNotFound, err
	}

	depth := infiniteDepth
	if hdr := r.Header.Get("Depth"); hdr != "" {
		depth = parseDepth(hdr)
		if depth == invalidDepth {
			return http.StatusBadRequest, errInvalidDepth
		}
	}
	pf, status, err := readPropfind(r.Body)
	if err != nil {
		return status, err
	}

	mw := multistatusWriter{w: w}

	walkFn := func(reqPath string, info FileInfo, err error) error {

		if err != nil {
			return err
		}
		var pstats []Propstat
		if pf.Propname != nil {
			pnames, err := propnames(ctx, fs, h.LockSystem[fs.User.ID], info)
			if err != nil {
				return err
			}
			pstat := Propstat{Status: http.StatusOK}
			for _, xmlname := range pnames {
				pstat.Props = append(pstat.Props, Property{XMLName: xmlname})
			}
			pstats = append(pstats, pstat)
		} else if pf.Allprop != nil {
			pstats, err = allprop(ctx, fs, h.LockSystem[fs.User.ID], info, pf.Prop)
		} else {
			pstats, err = props(ctx, fs, h.LockSystem[fs.User.ID], info, pf.Prop)
		}
		if err != nil {
			return err
		}
		href := path.Join(h.Prefix, reqPath)
		if href != "/" && info.IsDir() {
			href += "/"
		}
		return mw.write(makePropstatResponse(href, pstats))
	}

	walkErr := walkFS(ctx, fs, depth, reqPath, fi, walkFn)
	closeErr := mw.close()
	if walkErr != nil {
		return http.StatusInternalServerError, walkErr
	}
	if closeErr != nil {
		return http.StatusInternalServerError, closeErr
	}
	return 0, nil
}

func (h *Handler) handleProppatch(w http.ResponseWriter, r *http.Request, fs *filesystem.FileSystem) (status int, err error) {
	defer fs.Recycle()

	reqPath, status, err := h.stripPrefix(r.URL.Path, fs.User.ID)
	if err != nil {
		return status, err
	}
	release, status, err := h.confirmLocks(r, reqPath, "", fs)
	if err != nil {
		return status, err
	}
	defer release()

	ctx := r.Context()

	if exist, _ := isPathExist(ctx, fs, reqPath); !exist {
		return http.StatusNotFound, nil
	}
	patches, status, err := readProppatch(r.Body)
	if err != nil {
		return status, err
	}
	pstats, err := patch(ctx, fs, h.LockSystem[fs.User.ID], reqPath, patches)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	mw := multistatusWriter{w: w}
	writeErr := mw.write(makePropstatResponse(r.URL.Path, pstats))
	closeErr := mw.close()
	if writeErr != nil {
		return http.StatusInternalServerError, writeErr
	}
	if closeErr != nil {
		return http.StatusInternalServerError, closeErr
	}
	return 0, nil
}

func makePropstatResponse(href string, pstats []Propstat) *response {
	resp := response{
		Href:     []string{(&url.URL{Path: href}).EscapedPath()},
		Propstat: make([]propstat, 0, len(pstats)),
	}
	for _, p := range pstats {
		var xmlErr *xmlError
		if p.XMLError != "" {
			xmlErr = &xmlError{InnerXML: []byte(p.XMLError)}
		}
		resp.Propstat = append(resp.Propstat, propstat{
			Status:              fmt.Sprintf("HTTP/1.1 %d %s", p.Status, StatusText(p.Status)),
			Prop:                p.Props,
			ResponseDescription: p.ResponseDescription,
			Error:               xmlErr,
		})
	}
	return &resp
}

const (
	infiniteDepth = -1
	invalidDepth  = -2
)

// parseDepth maps the strings "0", "1" and "infinity" to 0, 1 and
// infiniteDepth. Parsing any other string returns invalidDepth.
//
// Different WebDAV methods have further constraints on valid depths:
//	- PROPFIND has no further restrictions, as per section 9.1.
//	- COPY accepts only "0" or "infinity", as per section 9.8.3.
//	- MOVE accepts only "infinity", as per section 9.9.2.
//	- LOCK accepts only "0" or "infinity", as per section 9.10.3.
// These constraints are enforced by the handleXxx methods.
func parseDepth(s string) int {
	switch s {
	case "0":
		return 0
	case "1":
		return 1
	case "infinity":
		return infiniteDepth
	}
	return invalidDepth
}

// http://www.webdav.org/specs/rfc4918.html#status.code.extensions.to.http11
const (
	StatusMulti               = 207
	StatusUnprocessableEntity = 422
	StatusLocked              = 423
	StatusFailedDependency    = 424
	StatusInsufficientStorage = 507
)

func StatusText(code int) string {
	switch code {
	case StatusMulti:
		return "Multi-Status"
	case StatusUnprocessableEntity:
		return "Unprocessable Entity"
	case StatusLocked:
		return "Locked"
	case StatusFailedDependency:
		return "Failed Dependency"
	case StatusInsufficientStorage:
		return "Insufficient Storage"
	}
	return http.StatusText(code)
}

var (
	errDestinationEqualsSource = errors.New("webdav: destination equals source")
	errDirectoryNotEmpty       = errors.New("webdav: directory not empty")
	errInvalidDepth            = errors.New("webdav: invalid depth")
	errInvalidDestination      = errors.New("webdav: invalid destination")
	errInvalidIfHeader         = errors.New("webdav: invalid If header")
	errInvalidLockInfo         = errors.New("webdav: invalid lock info")
	errInvalidLockToken        = errors.New("webdav: invalid lock token")
	errInvalidPropfind         = errors.New("webdav: invalid propfind")
	errInvalidProppatch        = errors.New("webdav: invalid proppatch")
	errInvalidResponse         = errors.New("webdav: invalid response")
	errInvalidTimeout          = errors.New("webdav: invalid timeout")
	errNoFileSystem            = errors.New("webdav: no file system")
	errNoLockSystem            = errors.New("webdav: no lock system")
	errNotADirectory           = errors.New("webdav: not a directory")
	errPrefixMismatch          = errors.New("webdav: prefix mismatch")
	errRecursionTooDeep        = errors.New("webdav: recursion too deep")
	errUnsupportedLockInfo     = errors.New("webdav: unsupported lock info")
	errUnsupportedMethod       = errors.New("webdav: unsupported method")
)
