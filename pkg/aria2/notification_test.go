package aria2

import (
	"testing"

	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/stretchr/testify/assert"
)

func TestNotifier_Notify(t *testing.T) {
	asserts := assert.New(t)
	notifier2 := &Notifier{}
	notifyChan := make(chan StatusEvent, 10)
	notifier2.Subscribe(notifyChan, "1")

	// 未订阅
	{
		notifier2.Notify([]rpc.Event{rpc.Event{Gid: ""}}, 1)
		asserts.Len(notifyChan, 0)
	}

	// 订阅
	{
		notifier2.Notify([]rpc.Event{{Gid: "1"}}, 1)
		asserts.Len(notifyChan, 1)
		<-notifyChan

		notifier2.OnBtDownloadComplete([]rpc.Event{{Gid: "1"}})
		asserts.Len(notifyChan, 1)
		<-notifyChan

		notifier2.OnDownloadStart([]rpc.Event{{Gid: "1"}})
		asserts.Len(notifyChan, 1)
		<-notifyChan

		notifier2.OnDownloadPause([]rpc.Event{{Gid: "1"}})
		asserts.Len(notifyChan, 1)
		<-notifyChan

		notifier2.OnDownloadStop([]rpc.Event{{Gid: "1"}})
		asserts.Len(notifyChan, 1)
		<-notifyChan

		notifier2.OnDownloadComplete([]rpc.Event{{Gid: "1"}})
		asserts.Len(notifyChan, 1)
		<-notifyChan

		notifier2.OnDownloadError([]rpc.Event{{Gid: "1"}})
		asserts.Len(notifyChan, 1)
		<-notifyChan
	}
}
