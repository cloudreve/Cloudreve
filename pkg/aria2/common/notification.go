package common

import (
	"sync"

	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
)

// Notifier aria2事件通知处理
type Notifier struct {
	Subscribes sync.Map
}

type CallbackFunc func(StatusEvent)

// Subscribe 订阅事件通知
func (notifier *Notifier) Subscribe(target chan StatusEvent, gid string) {
	notifier.Subscribes.Store(gid, target)
}

// Subscribe 订阅事件通知回调
func (notifier *Notifier) SubscribeCallback(callback CallbackFunc, gid string) {
	notifier.Subscribes.Store(gid, callback)
}

// Unsubscribe 取消订阅事件通知
func (notifier *Notifier) Unsubscribe(gid string) {
	notifier.Subscribes.Delete(gid)
}

// Notify 发送通知
func (notifier *Notifier) Notify(events []rpc.Event, status int) {
	for _, event := range events {
		if target, ok := notifier.Subscribes.Load(event.Gid); ok {
			msg := StatusEvent{
				GID:    event.Gid,
				Status: status,
			}

			if callback, ok := target.(CallbackFunc); ok {
				go callback(msg)
			} else {
				target.(chan StatusEvent) <- msg
			}

		}
	}
}

// OnDownloadStart 下载开始
func (notifier *Notifier) OnDownloadStart(events []rpc.Event) {
	notifier.Notify(events, Downloading)
}

// OnDownloadPause 下载暂停
func (notifier *Notifier) OnDownloadPause(events []rpc.Event) {
	notifier.Notify(events, Paused)
}

// OnDownloadStop 下载停止
func (notifier *Notifier) OnDownloadStop(events []rpc.Event) {
	notifier.Notify(events, Canceled)
}

// OnDownloadComplete 下载完成
func (notifier *Notifier) OnDownloadComplete(events []rpc.Event) {
	notifier.Notify(events, Complete)
}

// OnDownloadError 下载出错
func (notifier *Notifier) OnDownloadError(events []rpc.Event) {
	notifier.Notify(events, Error)
}

// OnBtDownloadComplete BT下载完成
func (notifier *Notifier) OnBtDownloadComplete(events []rpc.Event) {
	notifier.Notify(events, Complete)
}

// StatusEvent 状态改变事件
type StatusEvent struct {
	GID    string
	Status int
}
