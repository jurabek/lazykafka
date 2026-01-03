package types

// Notifier provides change notification via channels
type Notifier interface {
	NotifyChannel() <-chan struct{}
	Notify()
}
