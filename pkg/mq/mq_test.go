package mq

import (
	"github.com/cloudreve/Cloudreve/v3/pkg/aria2/rpc"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestPublishAndSubscribe(t *testing.T) {
	t.Parallel()
	asserts := assert.New(t)
	mq := NewMQ()

	// No subscriber
	{
		asserts.NotPanics(func() {
			mq.Publish("No subscriber", Message{})
		})
	}

	// One channel subscriber
	{
		topic := "One channel subscriber"
		msg := Message{TriggeredBy: "Tester"}
		notifier := mq.Subscribe(topic, 0)
		mq.Publish(topic, msg)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			wg.Done()
			msgRecv := <-notifier
			asserts.Equal(msg, msgRecv)
		}()
		wg.Wait()
	}

	// two channel subscriber
	{
		topic := "two channel subscriber"
		msg := Message{TriggeredBy: "Tester"}
		notifier := mq.Subscribe(topic, 0)
		notifier2 := mq.Subscribe(topic, 0)
		mq.Publish(topic, msg)
		wg := sync.WaitGroup{}
		wg.Add(2)
		go func() {
			wg.Done()
			msgRecv := <-notifier
			asserts.Equal(msg, msgRecv)
		}()
		go func() {
			wg.Done()
			msgRecv := <-notifier2
			asserts.Equal(msg, msgRecv)
		}()
		wg.Wait()
	}

	// two channel subscriber, one timeout
	{
		topic := "two channel subscriber, one timeout"
		msg := Message{TriggeredBy: "Tester"}
		mq.Subscribe(topic, 0)
		notifier2 := mq.Subscribe(topic, 0)
		mq.Publish(topic, msg)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			wg.Done()
			msgRecv := <-notifier2
			asserts.Equal(msg, msgRecv)
		}()
		wg.Wait()
	}

	// two channel subscriber, one unsubscribe
	{
		topic := "two channel subscriber, one unsubscribe"
		msg := Message{TriggeredBy: "Tester"}
		mq.Subscribe(topic, 0)
		notifier2 := mq.Subscribe(topic, 0)
		notifier := mq.Subscribe(topic, 0)
		mq.Unsubscribe(topic, notifier)
		mq.Publish(topic, msg)
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			wg.Done()
			msgRecv := <-notifier2
			asserts.Equal(msg, msgRecv)
		}()
		wg.Wait()

		select {
		case <-notifier:
			t.Error()
		default:
		}
	}
}

func TestAria2Interface(t *testing.T) {
	t.Parallel()
	asserts := assert.New(t)
	mq := NewMQ()
	var (
		OnDownloadStart    int
		OnDownloadPause    int
		OnDownloadStop     int
		OnDownloadComplete int
		OnDownloadError    int
	)
	l := sync.Mutex{}

	mq.SubscribeCallback("TestAria2Interface", func(message Message) {
		asserts.Equal("TestAria2Interface", message.TriggeredBy)
		l.Lock()
		defer l.Unlock()
		switch message.Event {
		case "1":
			OnDownloadStart++
		case "2":
			OnDownloadPause++
		case "5":
			OnDownloadStop++
		case "4":
			OnDownloadComplete++
		case "3":
			OnDownloadError++
		}
	})

	mq.OnDownloadStart([]rpc.Event{{"TestAria2Interface"}, {"TestAria2Interface"}})
	mq.OnDownloadPause([]rpc.Event{{"TestAria2Interface"}, {"TestAria2Interface"}})
	mq.OnDownloadStop([]rpc.Event{{"TestAria2Interface"}, {"TestAria2Interface"}})
	mq.OnDownloadComplete([]rpc.Event{{"TestAria2Interface"}, {"TestAria2Interface"}})
	mq.OnDownloadError([]rpc.Event{{"TestAria2Interface"}, {"TestAria2Interface"}})
	mq.OnBtDownloadComplete([]rpc.Event{{"TestAria2Interface"}, {"TestAria2Interface"}})

	time.Sleep(time.Duration(500) * time.Millisecond)

	asserts.Equal(2, OnDownloadStart)
	asserts.Equal(2, OnDownloadPause)
	asserts.Equal(2, OnDownloadStop)
	asserts.Equal(4, OnDownloadComplete)
	asserts.Equal(2, OnDownloadError)
}
